package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// ImageConfig holds the configs needed to build an image
type ImageConfig struct {
	Name       string
	Dockerfile string `yaml:",omitempty"`
	// whether to enable the no-cache option in docker build
	CachedBuild bool `yaml:"cache,omitempty" json:"cache"`
	// Additional tags to be applied to the image on top of the default 'latest'
	Tags []string `yaml:",omitempty"`

	Registry string `yaml:",omitempty"`
	Context  string `yaml:",omitempty"` // working directory, url etc.

	baseimage string

	id string
}

// Validate the image config.
func (ic *ImageConfig) Validate() error {

	if strings.Contains(ic.Name, ":") && len(ic.Tags) > 0 {
		return fmt.Errorf("cannot specify tags in name and tags")
	}

	return nil
}

// ReplaceTagVars replaces any instances of the placeholder found in tags with the supplied value.
func (ic *ImageConfig) ReplaceTagVars(placeholder, value string) {
	for i, tag := range ic.Tags {
		ic.Tags[i] = strings.Replace(tag, placeholder, value, -1)
	}
}

// DefaultRegistryPaths return the default (i.e. docker) registry paths
func (ic *ImageConfig) DefaultRegistryPaths() []string {
	paths := make([]string, len(ic.Tags)+1)
	paths[0] = ic.Name
	for i, tag := range ic.Tags {
		paths[i+1] = fmt.Sprintf("%s:%s", ic.Name, tag)
	}
	return paths
}

// CustomRegistryPaths returns the custom registry paths for this image.
func (ic *ImageConfig) CustomRegistryPaths() []string {
	if len(ic.Registry) == 0 {
		return []string{}
	}

	paths := make([]string, len(ic.Tags)+1)
	paths[0] = fmt.Sprintf("%s/%s", ic.Registry, ic.Name)
	for i, tag := range ic.Tags {
		paths[i+1] = fmt.Sprintf("%s/%s:%s", ic.Registry, ic.Name, tag)
	}
	return paths
}

func (ic *ImageConfig) RegistryPaths() []string {
	// if not registry specified return default docker registry
	if len(ic.Registry) == 0 {
		return ic.DefaultRegistryPaths()
	}
	return ic.CustomRegistryPaths()
}

/*// RegistryPath return the full image path to the registry
func (ic *ImageConfig) RegistryPath() string {
	if len(ic.Registry) == 0 {
		return ic.Name
	}
	return fmt.Sprintf("%s/%s", ic.Registry, ic.Name)
}*/

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
