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
}

// BuildCmds returns the command string that is passed in to bash -cex on the
// container.
func (cb *DockerRunConfig) BuildCmds() string {
	return strings.Join(cb.Commands, "\n") + "\n"
}
