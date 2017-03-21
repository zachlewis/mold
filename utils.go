package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
)

const dockerSockFile = "/var/run/docker.sock"

// parse git info from .git/HEAD to get name, branch and commit info.  If not found
// that item will be an empty string
func getRepoInfo(path string) (name, branchTag, lastCommit string) {
	name = filepath.Base(path)

	b, err := ioutil.ReadFile(filepath.Join(path, ".git/HEAD"))
	if err != nil {
		return
	}
	lp := strings.Split(string(bytes.TrimSuffix(b, []byte("\n"))), " ")

	switch len(lp) {
	case 2:
		pp := strings.Split(lp[1], "/")
		branchTag = pp[len(pp)-1]
		if cmt, err := ioutil.ReadFile(filepath.Join(path, ".git", lp[1])); err == nil {
			if len(cmt) > 7 {
				lastCommit = string(cmt[:8])
			}
		}

	case 1:
		if len(lp[0]) > 7 {
			lastCommit = lp[0][:8]
		}
	}

	return name, branchTag, lastCommit
}

// returns the name of the image.  it parses out the namespace and tag if provided
func nameFromImageName(imageName string) string {
	iparts := strings.Split(strings.Split(imageName, ":")[0], "/")
	if len(iparts) == 1 {
		return iparts[0]
	}
	return iparts[len(iparts)-1]
}

func assembleServiceContainers(bc *BuildConfig) []*ContainerConfig {
	bcs := make([]*ContainerConfig, len(bc.Services))
	for i, b := range bc.Services {
		cc := DefaultContainerConfig(b.Image)
		cc.Container.Cmd = b.Commands
		cc.Container.Env = b.Environment
		bcs[i] = cc
	}
	return bcs
}

// assembleBuildContainers assembles container configs from user provided build config
func assembleBuildContainers(bc *BuildConfig) ([]*ContainerConfig, error) {
	bconts := make([]*ContainerConfig, len(bc.Build))
	for i, b := range bc.Build {
		cc := DefaultContainerConfig(b.Image)
		cc.Container.WorkingDir = b.Workdir

		exposedPorts, portBindings, err := nat.ParsePortSpecs(b.Ports)
		if err != nil {
			return nil, err
		}
		cc.Container.ExposedPorts = exposedPorts
		cc.Host.PortBindings = portBindings

		cc.Container.Volumes = map[string]struct{}{b.Workdir: struct{}{}}
		cc.Container.Cmd = []string{b.Shell, "-cex", b.BuildCmds()}
		cc.Container.Env = b.Environment
		src := bc.Context
		if runtime.GOOS == "windows" {
			src = toDockerWinPath(src)
		}
		cc.Host.Mounts = []mount.Mount{
			mount.Mount{Target: b.Workdir, Source: src, Type: mount.TypeBind},
		}
		bconts[i] = cc

		// Mount docker.sock in container if requested.
		if bc.AllowDockerAccess {
			bconts[i].Container.Volumes[dockerSockFile] = struct{}{}
			bconts[i].Host.Mounts = append(bconts[i].Host.Mounts,
				mount.Mount{Target: dockerSockFile, Source: dockerSockFile, Type: mount.TypeBind})
		}
	}
	return bconts, nil
}

// Merges errors together
func mergeErrors(err1, err2 error) error {
	if err1 == nil {
		return err2
	} else if err2 == nil {
		return err1
	} else {
		return fmt.Errorf("%s\n%s", err1, err2)
	}
}

func readBuildConfig(bldfile string) (*BuildConfig, error) {
	d, err := ioutil.ReadFile(bldfile)
	if err == nil {
		return NewBuildConfig(d)
	}
	return nil, err
}

func tarDirectory(srcPath string) (io.ReadCloser, error) {
	var excludes []string
	includes := []string{"."}

	tarOpts := &archive.TarOptions{
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
		Compression:     archive.Gzip,
		NoLchown:        true,
	}
	return archive.TarWithOptions(srcPath, tarOpts)
}

// parseTarget parses user supplied target to get the lifecycle phase and
// a sub phase is specified
func parseTarget(target string) (lcStep LifeCyclePhase, sub string) {
	if target == "" {
		return
	}

	pp := strings.Split(target, "/")
	switch len(pp) {
	case 1:
		lcStep = LifeCyclePhase(pp[0])
	default:
		lcStep = LifeCyclePhase(pp[0])
		sub = strings.Join(pp[1:], "/")
	}
	return
}

func toDockerWinPath(p string) string {
	p = strings.Replace(p, `\`, "/", -1)
	if !strings.HasPrefix(p, "/") {
		p = "//" + p
	}
	p = strings.Replace(p, ":", "", 1)
	return p
}

func printUsage() {
	fmt.Printf(`
mold [ options ]

Mold is a tool to perform testing, building, packaging and publishing of
applications completely using Docker.  Application code is tested and built in a
Docker container following by the building for Docker images and publishing to a
registry, all controlled via a single configuration file.

Options:

  -version  Show version

  -var      Show value of vairable specified in the configuration file  (default: NA)

  -uri      Docker URI          (default: %s)

  -f        Configuration file  (default: %s)

  -t        Target to build     (default: all)

            build       Only perform the build phase.

            artifacts   Only generate artifacts.  Specific artifacts can be built
                        using artifacts/<image_name> as the target where
                        <image_name> would be that as specified in your
                        configuration.

            publish     Only publish artifacts.  Specific artifacts can be published
                        using publish/<image_name> as the target where
                        <image_name> would be that as specified in your
                        configuration.

`, *dockerURI, *buildFile)
}
