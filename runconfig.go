package main

import (
	"strings"
)

// DockerRunConfig holds the config to run a container
type DockerRunConfig struct {
	Image       string   // Docker image to use for code build
	Commands    []string // Commands to run in the container
	Workdir     string   // Working directory in the container
	Environment []string
	Save        bool      `yaml:",omitempty"` // do not remove container after completion
	Shell       string    `yaml:",omitempty"`
	Ports       []string  `yaml:",omitempty"` // a quoted list of port mappings
	ImgCache    *ImgCache `yaml:",omitempty"`
}

type ImgCache struct {
	Registry string
	Name     string
	Tag      string
}

func (ic *ImgCache) IsEnabled() bool {
	return ic != nil && len(ic.Registry) > 0 && len(ic.Name) > 0
}

func (ic *ImgCache) IsTagSet() bool {
	return ic != nil && len(ic.Tag) > 0
}

// BuildCmds returns the command string that is passed in to bash -cex on the
// container.
func (cb *DockerRunConfig) BuildCmds() string {
	return strings.Join(cb.Commands, "\n") + "\n"
}
