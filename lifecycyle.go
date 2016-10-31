package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	// lifeCycleInit is where configs are read in and validated
	lifeCycleConfigure byte = iota
	// lifeCycleSetup sets up additional containers that need to be run for the build
	lifeCycleSetup
	// lifeCycleBuild is where the user defined work is performed in the container
	lifeCycleBuild
	// lifeCyleArtifacts builds specified docker images
	lifeCyleArtifacts
	// lifeCyclePublish pushes the docker images up to a registry
	lifeCyclePublish
	// lifeCycleTeardown cleans up resources created during the build.
	lifeCycleTeardown
)

// Worker perform all work for a given job.  This would be implemented
// based on the backend used - in the current case docker.
type Worker interface {
	Configure(*BuildConfig) error      // Initialize underlying needed structs
	Setup() error                      // Statisfy deps needed for the build
	Build() error                      // Build required data to be package
	GenerateArtifacts(...string) error // Package data to an artifact.
	Publish(...string) error           // Publish the generated artifacts
	Teardown() error
}

// LifeCycle manages the complete lifecyle
type LifeCycle struct {
	worker Worker
	cfg    *BuildConfig
	log    io.Writer
}

// NewLifeCycle with stdout as the logger with the provided worker
func NewLifeCycle(worker Worker) *LifeCycle {
	return &LifeCycle{worker: worker, log: os.Stdout}
}

// Run the complete lifecyle
func (lc *LifeCycle) Run(cfg *BuildConfig) error {

	err := lc.worker.Configure(cfg)
	if err != nil {
		return err
	}
	lc.cfg = cfg
	lc.printStartSummary()

	if err = lc.worker.Setup(); err == nil {
		if err = lc.worker.Build(); err == nil {
			if err = lc.worker.GenerateArtifacts(); err == nil {
				if lc.shouldPublishArtifacts() {
					err = lc.worker.Publish()
				} else {
					lc.log.Write([]byte("[publish] Not publishing. Criteria not met.\n"))
				}
			}
		}
	}
	if e := lc.worker.Teardown(); e != nil {
		log.Printf("ERR [Teardown] %v", e)
	}

	return err
}

func (lc *LifeCycle) shouldPublishArtifacts() bool {
	arts := lc.cfg.Artifacts
	for _, p := range arts.Publish {
		switch p {
		case "*", lc.cfg.BranchTag:
			return true
		}
	}
	return false
}

// RunTarget runs a specified target in the lifecyle
func (lc *LifeCycle) RunTarget(cfg *BuildConfig, t byte) error {
	var err error
	switch t {
	case lifeCycleBuild:
		if err = lc.worker.Configure(cfg); err == nil {
			if err = lc.worker.Setup(); err == nil {
				err = lc.worker.Build()
			}
		}
		if e := lc.worker.Teardown(); e != nil {
			log.Printf("ERR [Teardown] %v", e)
		}

	case lifeCyleArtifacts:
		if err = lc.worker.Configure(cfg); err == nil {
			err = lc.worker.GenerateArtifacts()
		}

	case lifeCyclePublish:
		if err = lc.worker.Configure(cfg); err == nil {
			err = lc.worker.Publish()
		}

	default:
		err = fmt.Errorf("invalid target: %d", t)

	}
	return err
}

func (lc *LifeCycle) printStartSummary() {
	c := lc.cfg
	lc.log.Write([]byte(fmt.Sprintf(`
Name       : %s
Branch/Tag : %s
Repo       : %s

Builds     : %d
Artifacts  : %d

`, c.Name, c.BranchTag, c.RepoURL, len(c.Build), len(c.Artifacts.Images))))
}
