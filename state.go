package main

import "fmt"

// ContainerType is the type of container.  This is used to govern the containers
// lifecycle
type ContainerType byte

const (
	// ServiceContainerType is a service container which is transient
	ServiceContainerType ContainerType = iota
	// BuildContainerType is a build container
	BuildContainerType
)

// containerState represents the current state of a given container
type containerState struct {
	*ContainerConfig

	Type   ContainerType // service or build
	status string        // build status of the container
	done   bool          // container execution completed
	save   bool          // keep the container after run completes
	cache  *cache
}

type cache struct {
	Name string
	Tag  string
}

func (ic *cache) IsSet() bool {
	return ic != nil && len(ic.Name) > 0 && len(ic.Tag) > 0
}

func (ic *cache) ToString() string {
	if ic == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", ic.Name, ic.Tag)
}

func (cs *containerState) Status() string {
	if cs.state != nil {
		if cs.state.ExitCode != 0 {
			return "failed"
		}
		return "success"
	}
	return cs.status
}

type containerStates []*containerState

func (cs containerStates) Get(id string) *containerState {
	for i, v := range cs {
		if v.ID() == id {
			return cs[i]
		}
	}
	return nil
}
