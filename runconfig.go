package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
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
	EnvFiles    []string `yaml:"env_file,omitempty"` // files with environment variables
}

// BuildCmds returns the command string that is passed in to bash -cex on the
// container.
func (cb *DockerRunConfig) BuildCmds() string {
	return strings.Join(cb.Commands, "\n") + "\n"
}

// GetEnvStrings returns a list of env strings collected from the EnvFiles and the Environment.
func (cb *DockerRunConfig) GetEnvStrings() ([]string, error) {
	vars := []string{}
	for _, ef := range cb.EnvFiles {
		efvars, err := parseEnvFile(ef)
		if err != nil {
			return nil, err
		}
		vars = append(vars, efvars...)
	}
	for _, ef := range cb.Environment {
		efvars, err := formatEnvVar([]byte(ef))
		if err != nil {
			return nil, err
		}
		vars = append(vars, efvars)
	}
	return vars, nil
}

func parseEnvFile(filename string) ([]string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var ls []string
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		evar := scanner.Bytes()
		l, err := formatEnvVar(evar)
		if err != nil {
			return nil, err
		}
		if len(l) > 0 {
			ls = append(ls, l)
		}
	}
	return ls, nil
}

func formatEnvVar(evar []byte) (string, error) {
	if !utf8.Valid(evar) {
		return "", fmt.Errorf("invalid utf8 char found in env var: %v", evar)
	}
	evarStr := strings.TrimLeftFunc(string(evar), unicode.IsSpace)

	if len(evarStr) > 0 && !strings.HasPrefix(evarStr, "#") {
		return evarStr, nil
	}
	return "", nil
}
