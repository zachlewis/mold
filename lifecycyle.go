package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LifeCyclePhase represents a phase in the lifecycle
type LifeCyclePhase string

const (
	lifeCycleConfigure LifeCyclePhase = "configure" // lifeCycleInit is where configs are read in and validated
	lifeCycleSetup     LifeCyclePhase = "setup"     // lifeCycleSetup sets up additional containers that need to be run for the build
	lifeCycleBuild     LifeCyclePhase = "build"     // lifeCycleBuild is where the user defined work is performed in the container
	lifeCyleArtifacts  LifeCyclePhase = "artifacts" // lifeCyleArtifacts builds specified docker images
	lifeCyclePublish   LifeCyclePhase = "publish"   // lifeCyclePublish pushes the docker images up to a registry
	lifeCycleTeardown  LifeCyclePhase = "teardown"  // lifeCycleTeardown cleans up resources created during the build.
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
	Abort() error
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

// whether to publish the image based on the branch/tag
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

// Abort the lifecyle ending it.
func (lc *LifeCycle) Abort() error {
	return lc.worker.Abort()
}

// RunTarget runs a specified target in the lifecyle
func (lc *LifeCycle) RunTarget(cfg *BuildConfig, target LifeCyclePhase, args ...string) error {
	var err error
	switch target {
	case lifeCycleBuild:
		if err = lc.worker.Configure(cfg); err == nil {
			if err = lc.worker.Setup(); err == nil {
				err = lc.worker.Build()
			}
		}
		if e := lc.worker.Teardown(); e != nil {
			log.Printf("ERR [%s] %v", lifeCycleTeardown, e)
		}

	case lifeCyleArtifacts:
		if err = lc.worker.Configure(cfg); err == nil {
			err = lc.worker.GenerateArtifacts(args...)
		}

	case lifeCyclePublish:
		if err = lc.worker.Configure(cfg); err == nil {
			err = lc.worker.Publish(args...)
		}

	default:
		err = fmt.Errorf("invalid target: %s", target)

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
