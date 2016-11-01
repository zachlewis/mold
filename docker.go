package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// ContainerConfig holds all configs needed to run the docker container
type ContainerConfig struct {
	Name      string // name of the running container
	Container *container.Config
	Host      *container.HostConfig
	Network   *network.NetworkingConfig

	id    string
	state *types.ContainerState
}

// ID returns the id as set by docker. This gets set once the container has
// started
func (cc *ContainerConfig) ID() string {
	return cc.id
}

// IsRunning reports if the container is running per the local state.
func (cc *ContainerConfig) IsRunning() bool {
	return cc.state.Running
}

// DefaultContainerConfig contains just the image name
func DefaultContainerConfig(imageName string) *ContainerConfig {
	return &ContainerConfig{
		Container: &container.Config{Image: imageName, Volumes: map[string]struct{}{}},
		Host:      &container.HostConfig{Binds: []string{}, Mounts: []mount.Mount{}},
		Network:   &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}},
	}
}

// ImageConfig holds the configs needed to build an image
type ImageConfig struct {
	Name       string
	Dockerfile string
	Registry   string
	Context    string // working directory, url etc.

	baseimage string

	id string
}

// RegistryPath return the full path to the registry
func (ic *ImageConfig) RegistryPath() string {
	if len(ic.Registry) == 0 {
		return ic.Name
	}
	return fmt.Sprintf("%s/%s", ic.Registry, ic.Name)
}

// BaseImage reads the baseimage from the dockerfile if not caches otherwise
// returns the cached copy
func (ic *ImageConfig) BaseImage() (string, error) {
	if len(ic.baseimage) > 0 {
		return ic.baseimage, nil
	}

	b, err := ioutil.ReadFile(ic.Dockerfile)
	if err == nil {
		sb := string(b)
		for _, s := range strings.Split(sb, "\n") {
			if strings.HasPrefix(s, "FROM ") {
				p := strings.Split(s, " ")
				ic.baseimage = strings.TrimSpace(p[len(p)-1])
				return ic.baseimage, nil
			}
		}
		err = fmt.Errorf("FROM entry not found: %s", ic.Dockerfile)
	}
	return "", err
}

// Docker provides a wrapper to perform rudamentary docker operations
type Docker struct {
	cli *client.Client
}

// Client returns the raw docker client
func (dkr *Docker) Client() *client.Client {
	return dkr.cli
}

// ImageAvailableLocally returns true if the image is locally available
func (dkr *Docker) ImageAvailableLocally(imageName string) bool {
	if _, _, err := dkr.cli.ImageInspectWithRaw(context.Background(), imageName); err == nil {
		return true
	}
	return false
}

// StartContainer creates and starts a container with the given config updating
// the state of the ContainerConfig.  It also pulls the base image if not locally
// available. This is a non-blocking call.
func (dkr *Docker) StartContainer(cc *ContainerConfig, wr io.Writer, prefix string) error {
	if !dkr.ImageAvailableLocally(cc.Container.Image) {
		if err := dkr.PullImage(cc.Container.Image, wr, prefix); err != nil {
			return err
		}
	}

	c, err := dkr.cli.ContainerCreate(context.Background(), cc.Container, cc.Host, cc.Network, cc.Name)
	if err != nil {
		return err
	}
	cc.id = c.ID

	if err = dkr.cli.ContainerStart(context.Background(), cc.id, types.ContainerStartOptions{}); err == nil {

		var cont types.ContainerJSON
		if cont, err = dkr.cli.ContainerInspect(context.Background(), cc.id); err == nil {
			cc.state = cont.State
			if len(cc.Name) == 0 {
				cc.Name = cont.Name[1:]
			}
		}
	}
	return err
}

// TailLogs tail container logs to the given writer
// Attach stdout and stderr of the container to stdout
func (dkr *Docker) TailLogs(containerID string, wr io.Writer, prefix string) error {
	opts := types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Follow:     true,
		Timestamps: true,
	}
	r, err := dkr.cli.ContainerLogs(context.Background(), containerID, opts)
	if err != nil {
		return err
	}
	defer r.Close()
	buf := bufio.NewReader(r)

	for {
		var b []byte
		if b, err = buf.ReadBytes('\n'); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		// First 8 bytes are the stream type
		// Append the specified prefix
		wr.Write([]byte(prefix + " "))
		// Remove the date from the timestamp
		wr.Write(append(append(b[19:32], ' '), b[39:]...))
	}
	return err
}

// RemoveContainer removes a container
func (dkr *Docker) RemoveContainer(containerID string, force bool) error {
	options := types.ContainerRemoveOptions{Force: force}
	return dkr.cli.ContainerRemove(context.Background(), containerID, options)
}

// BuildImage builds a docker images based on the config and writes the log out
// to the the specified Writer.  This is a blocking call.
func (dkr *Docker) BuildImage(ic *ImageConfig, logWriter io.Writer, prefix string) error {
	bldCxt, err := tarDirectory(ic.Context, nil)
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: ic.Dockerfile,
		Tags:       []string{ic.Name},
		Remove:     true, // remove intermediate images
	}
	if len(ic.Registry) > 0 {
		opts.Tags = append(opts.Tags, ic.RegistryPath())
	}

	rsp, err := dkr.cli.ImageBuild(context.Background(), bldCxt, opts)
	if err == nil {
		defer rsp.Body.Close()
		buf := bufio.NewReader(rsp.Body)

		for {
			var b []byte
			if b, err = buf.ReadBytes('\n'); err != nil {
				if err == io.EOF {
					err = nil
				}
				break
			}

			var m map[string]string
			if err = json.Unmarshal(b, &m); err != nil {
				logWriter.Write([]byte(fmt.Sprintf("%s ERR %s %s\n", prefix, err, b)))
				break
			}

			//if v, ok := m["stream"]; ok {
			logWriter.Write([]byte(prefix + " "))
			logWriter.Write([]byte(m["stream"]))
			//}
		}
	}
	return err
}

// RemoveImage locally from the host
func (dkr *Docker) RemoveImage(imageID string, force bool) error {
	options := types.ImageRemoveOptions{Force: force}
	_, err := dkr.cli.ImageRemove(context.Background(), imageID, options)
	return err
}

// CreateNetwork creates a bridge network
func (dkr *Docker) CreateNetwork(name string) (string, error) {
	opts := types.NetworkCreate{Driver: "bridge", CheckDuplicate: true}
	rsp, err := dkr.cli.NetworkCreate(context.Background(), name, opts)
	if err != nil {
		return "", err
	}
	return rsp.ID, nil
}

// RemoveNetwork removes the network from the host
func (dkr *Docker) RemoveNetwork(networkID string) error {
	return dkr.cli.NetworkRemove(context.Background(), networkID)
}

// PushImage pushes a local docker image up to a registry
func (dkr *Docker) PushImage(imageRef string, logWriter io.Writer) error {
	opts := types.ImagePushOptions{}
	rsp, err := dkr.cli.ImagePush(context.Background(), imageRef, opts)
	if err != nil {
		return err
	}
	defer rsp.Close()
	buf := bufio.NewReader(rsp)

	for {
		var b []byte
		if b, err = buf.ReadBytes('\n'); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		var m map[string]string
		if err = json.Unmarshal(b, &m); err != nil {
			break
		}
		logWriter.Write([]byte(m["status"]))
	}

	return err
}

// PullImage pulls a remote image from a registry down locally
func (dkr *Docker) PullImage(imageRef string, logWriter io.Writer, prefix string) error {
	opts := types.ImagePullOptions{}
	rsp, err := dkr.cli.ImagePull(context.Background(), imageRef, opts)
	if err != nil {
		return err
	}
	defer rsp.Close()
	buf := bufio.NewReader(rsp)

	logWriter.Write([]byte(fmt.Sprintf("%s Pulling image: %s\n", prefix, imageRef)))
	for {
		var b []byte
		if b, err = buf.ReadBytes('\n'); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		b = b[:len(b)-1]

		var m imgPullProgress
		if err = json.Unmarshal(b, &m); err != nil {
			break
		}

		if m.ProgressDetail.Current == m.ProgressDetail.Total && m.ProgressDetail.Total > 0 {
			//fmt.Printf("%s ... %s %d bytes\n", m.Status, m.ID, m.ProgressDetail.Total)
			logWriter.Write([]byte(fmt.Sprintf("%s %s: %s %d bytes\n", prefix, m.Status, m.ID, m.ProgressDetail.Total)))
		}

	}
	if err == nil {
		logWriter.Write([]byte(fmt.Sprintf("%s Pulled image: %s\n", prefix, imageRef)))
	}

	return err
}

type imgPullProgressDetail struct {
	Current  int
	Total    int
	Progress string
}

func (ipd *imgPullProgressDetail) Percent() float64 {
	return float64(ipd.Current) / float64(ipd.Total) * 100
}

type imgPullProgress struct {
	Status         string
	ID             string
	ProgressDetail imgPullProgressDetail
}
