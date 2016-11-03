package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

// default timeout when trying to stop a container.
const defaultStopTimeout int = 5

var errAborted = fmt.Errorf("aborted")

// DockerWorker performs a docker based build per the config
type DockerWorker struct {
	mu sync.Mutex

	cfg   *BuildConfig    // overall buildconfig
	sc    containerStates // service containers
	bc    containerStates // build containers
	netID string          // network id to connect all containers to

	dkr *Docker // docker helper client

	done    chan bool // when all builds are completed
	abort   chan bool // cancelled channel
	aborted bool      // whether the worker has begun shutdown

	log io.Writer
}

// NewDockerWorker instantiates a new worker. If no client is provided and env.
// based client is used.
func NewDockerWorker(dcli *Docker) (d *DockerWorker, err error) {
	d = &DockerWorker{dkr: dcli, log: os.Stdout, abort: make(chan bool, 1)}

	if d.dkr == nil {
		d.dkr, err = NewDocker("")
	}
	return
}

// Configure the job. This converts the BuildConfig to the docker required
// datastructure
func (bld *DockerWorker) Configure(cfg *BuildConfig) error {
	bld.mu.Lock()
	defer bld.mu.Unlock()
	bld.cfg = cfg

	// Build service container contfigs
	sc := assembleServiceContainers(cfg)
	bld.sc = make([]*containerState, len(sc))
	for i, s := range sc {
		// Initialize state
		cs := &containerState{ContainerConfig: s, Type: ServiceContainerType}
		cs.Name = nameFromImageName(s.Container.Image)
		// Attach network
		cs.Network = bld.defaultNetConfig()
		bld.sc[i] = cs
	}
	//time.Now().Format(time.RFC3339)
	// Build build container configs
	bc := assembleBuildContainers(cfg)
	bld.bc = make([]*containerState, len(bc))
	for i, s := range bc {
		cs := &containerState{ContainerConfig: s, Type: BuildContainerType}
		//cs.Name = fmt.Sprintf("%s-%d", bld.cfg.Name, i)

		cs.Name = fmt.Sprintf("%s-%d-%d", bld.cfg.Name, i, time.Now().UnixNano())
		cs.Network = bld.defaultNetConfig()
		bld.bc[i] = cs
	}
	return nil
}

func (bld *DockerWorker) defaultNetConfig() *network.NetworkingConfig {
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			bld.cfg.Name: &network.EndpointSettings{
				NetworkID: bld.netID,
			},
		},
	}
}

// GenerateArtifacts builds docker images
func (bld *DockerWorker) GenerateArtifacts(names ...string) error {
	if len(names) == 0 {
		for _, a := range bld.cfg.Artifacts.Images {
			if bld.aborted {
				return errAborted
			}
			bld.log.Write([]byte(fmt.Sprintf("[artifacts/%s] Building\n", a.Name)))
			if err := bld.dkr.BuildImage(&a, bld.log, fmt.Sprintf("[artifacts/%s]", a.Name)); err != nil {
				return err
			}
			bld.log.Write([]byte(fmt.Sprintf("[artifacts/%s] DONE\n", a.Name)))
		}
	} else {
		for _, name := range names {
			if bld.aborted {
				return errAborted
			}
			a := bld.cfg.Artifacts.GetImage(name)
			if a == nil {
				return fmt.Errorf("no such artifact: %s", name)
			}
			bld.log.Write([]byte(fmt.Sprintf("[artifacts/%s] Building\n", a.Name)))
			if err := bld.dkr.BuildImage(a, bld.log, fmt.Sprintf("[artifacts/%s]", a.Name)); err != nil {
				return err
			}
			bld.log.Write([]byte(fmt.Sprintf("[artifacts/%s] DONE\n", a.Name)))
		}
	}
	return nil
}

// RemoveArtifacts removes all local artifacts it as definted in the config
func (bld *DockerWorker) RemoveArtifacts() error {
	var err error
	for _, a := range bld.cfg.Artifacts.Images {
		err = mergeErrors(err, bld.dkr.RemoveImage(a.Name, true))
	}
	return err
}

// Publish the artifact/s based on the config
func (bld *DockerWorker) Publish(names ...string) error {
	if len(names) == 0 {
		for _, v := range bld.cfg.Artifacts.Images {
			if bld.aborted {
				return errAborted
			}
			if err := bld.dkr.PushImage(v.RegistryPath(), os.Stdout); err != nil {
				return err
			}
		}
	} else {
		for _, name := range names {
			if bld.aborted {
				return errAborted
			}
			a := bld.cfg.Artifacts.GetImage(name)
			if a == nil {
				return fmt.Errorf("no such artifact: %s", name)
			}
			if err := bld.dkr.PushImage(a.RegistryPath(), os.Stdout); err != nil {
				return err
			}
		}
	}
	return nil
}

// Build starts the build.  This is a blocking call. index defines one or more
// build steps to run.  They are in the order as seen in teh config. If no index
// is provided all builds are run
func (bld *DockerWorker) Build() error {
	done, err := bld.StartBuildAsync(true)
	if err != nil {
		return err
	}

	select {
	case <-done:
		for _, b := range bld.bc {
			if b.status != "success" {
				err = mergeErrors(err, fmt.Errorf("build failed: %s %s", b.Name, b.Container.Image))
			}
		}

	case <-bld.abort:
		bld.log.Write([]byte("[build] Aborting...\n"))
		if e := bld.stopBuildContainer(); e != nil {
			bld.log.Write([]byte("ERR Stopping build containers:" + e.Error() + "\n"))
		}
	}

	return err
}

func (bld *DockerWorker) stopBuildContainer() error {
	var err error
	for _, bc := range bld.bc {
		bld.log.Write([]byte("[build] Stopping container: " + bc.ID() + "\n"))
		err = mergeErrors(err, bld.dkr.StopContainer(bc.ID(), time.Duration(defaultStopTimeout)*time.Second))
	}
	return err
}

// Abort cancels a running build
func (bld *DockerWorker) Abort() error {
	bld.mu.Lock()
	bld.aborted = true
	bld.mu.Unlock()

	bld.abort <- true
	return nil
}

// Setup sets up services needed to perform the build.  These are additional containers
// that are spun up.  If any error occurs the whole build will bail out
func (bld *DockerWorker) Setup() error {
	var err error
	if bld.netID, err = bld.dkr.CreateNetwork(bld.cfg.Name); err != nil {
		return err
	}

	bld.log.Write([]byte(fmt.Sprintf("[configure/network/%s] Created %s\n", bld.cfg.Name, bld.netID)))
	for i, cs := range bld.sc {
		if err := bld.dkr.StartContainer(bld.sc[i].ContainerConfig, bld.log, fmt.Sprintf("[setup/service/%s]", cs.Name)); err != nil {
			return err
		}
		bld.log.Write([]byte(fmt.Sprintf("[setup/service/%s] Started %s\n", cs.Name, cs.Container.Image)))
	}
	return nil
}

// StartBuildAsync starts the build container/s
func (bld *DockerWorker) StartBuildAsync(tailLog bool) (chan bool, error) {

	bld.done = make(chan bool)
	go bld.watchBuild()

	for i, cs := range bld.bc {
		err := bld.dkr.StartContainer(bld.bc[i].ContainerConfig, bld.log, "")
		if err == nil {
			os.Stdout.Write([]byte(fmt.Sprintf("[build/%s] Started %s\n", cs.Name, cs.Container.Image)))
			if cs.Type == BuildContainerType && tailLog {
				go func(prefix string) {
					// wait otherwise docker may return a 404
					<-time.After(1 * time.Second)
					if e := bld.dkr.TailLogs(cs.ID(), bld.log, prefix); e != nil {
						log.Println("ERR Failed to tail log", e)
					}
				}(fmt.Sprintf("[build/%s]", cs.Name))
			}
			continue
		}
		return bld.done, err
	}

	return bld.done, nil
}

// Teardown stops and removes all services spun up before the build as part of cleanup
func (bld *DockerWorker) Teardown() error {
	var err error
	for _, cs := range bld.sc {
		e := bld.dkr.RemoveContainer(cs.ID(), true)
		err = mergeErrors(err, e)
	}
	err = mergeErrors(err, bld.dkr.RemoveNetwork(bld.netID))
	return err
}

// TODO: add locking???
// markContainerDone marks the container as done.  Return if all the build containers have completed
func (bld *DockerWorker) markContainerDone(id, status string, state *types.ContainerState) bool {
	for i, v := range bld.bc {
		if v.ID() == id {
			bld.mu.Lock()
			bld.bc[i].done = true

			if len(status) > 0 {
				bld.bc[i].status = status
			}
			if state != nil {
				bld.bc[i].state = state
			} else {
				bld.bc[i].state = &types.ContainerState{Running: false}
			}
			bld.mu.Unlock()
			bld.log.Write([]byte(fmt.Sprintf("[build/%s] DONE\n", v.Name)))
		}
	}
	// check if all builds are done
	for _, v := range bld.bc {
		if !v.done {
			return false
		}
	}
	bld.done <- true
	return true
}

func (bld *DockerWorker) watchBuild() {
	cli := bld.dkr.Client()
	msgCh, errCh := cli.Events(context.Background(), types.EventsOptions{})
	for {
		select {
		case msg := <-msgCh:

			switch msg.Action {
			case "destroy":
				// Check if we are interested in this container
				if c := bld.bc.Get(msg.Actor.ID); c != nil {
					// Breakout if the whole build is done.  This does not update the status
					// and is there more so the build doesn't block forever in case of failures
					if bld.markContainerDone(msg.Actor.ID, "", nil) {
						return
					}
				}

			case "die", "kill", "stop":
				// Check if we are interested in this container
				if c := bld.bc.Get(msg.Actor.ID); c != nil {
					var (
						status string
						state  types.ContainerState
					)
					if cj, err := cli.ContainerInspect(context.Background(), msg.Actor.ID); err == nil {
						//log.Printf("DOCKER %s %+v", cj.Image, cj.State)
						if cj.State.ExitCode != 0 {
							status = "failed"
						} else {
							status = "success"
						}
						state = *cj.State
					} else {
						status = msg.Action
					}
					// breakout if the whole build is done
					if bld.markContainerDone(msg.Actor.ID, status, &state) {
						return
					}
				}
			}

		case err := <-errCh:
			log.Println("ERR", err)

		}
	}
}
