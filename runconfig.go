package main

import (
	"strings"
)

// DockerRunConfig holds the config to run a container
type DockerRunConfig struct {
	Image       string   // Docker image to use for code build
	Commands    []string // Commands to run in the container
	Workdir     string   // Working directory in the container
	Environment []string `yaml:",omitempty"`
	Volumes     []string `yaml:"volumes,omitempty"`
	Save        bool     `yaml:",omitempty"` // do not remove container after completion
	Shell       string   `yaml:",omitempty"`
	Ports       []string `yaml:",omitempty"` // a quoted list of port mappings
	Cache       bool     `yaml:",omitempty"`
	Name        string   `yaml:",omitempty"`
	CleanUp     bool     `yaml:",omitempty"`
	File        string   `yaml:"file,omitempty"` // a file with environment variables
}

// BuildCmds returns the command string that is passed in to bash -cex on the
// container.
func (cb *DockerRunConfig) BuildCmds() string {
	return strings.Join(cb.Commands, "\n") + "\n"
}
