package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// ImageConfig holds the configs needed to build an image
type ImageConfig struct {
	Name        string
	Dockerfile  string `yaml:",omitempty"`
	CachedBuild bool   `yaml:"cache,omitempty" json:"cache"` // whether to enable the no-cache option in docker build

	Registry string `yaml:",omitempty"`
	Context  string `yaml:",omitempty"` // working directory, url etc.

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
