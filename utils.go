package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v1"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
)

const dockerSockFile = "/var/run/docker.sock"

// initializeMoldConfig is called with the -init flag. It creates a new config at the root
// of the project if one is not there.
func initializeMoldConfig(dirname string) error {
	cpath := filepath.Join(dirname, defaultBuildConfigName)
	if _, err := os.Stat(cpath); err == nil {
		return fmt.Errorf("%s already exists!", defaultBuildConfigName)
	}

	apath, err := filepath.Abs(dirname)
	if err != nil {
		return err
	}
	name := filepath.Base(apath)

	mc := DefaultMoldConfig(name)
	b, err := yaml.Marshal(mc)
	if err != nil {
		return err
	}

	fh, err := os.Create(cpath)
	if err != nil {
		return err
	}

	_, err = fh.Write(b)
	return err
}

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

func assembleServiceContainers(mc *MoldConfig) []*ContainerConfig {
	bcs := make([]*ContainerConfig, len(mc.Services))
	for i, b := range mc.Services {
		cc := DefaultContainerConfig(b.Image)
		cc.Container.Cmd = b.Commands
		cc.Container.Env = b.Environment
		bcs[i] = cc
	}
	return bcs
}

// assembleBuildContainers assembles container configs from user provided build config
func assembleBuildContainers(mc *MoldConfig) ([]*ContainerConfig, error) {
	bconts := make([]*ContainerConfig, len(mc.Build))
	for i, b := range mc.Build {
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
		src := mc.Context
		if runtime.GOOS == "windows" {
			src = toDockerWinPath(src)
		}
		cc.Host.Mounts = []mount.Mount{
			mount.Mount{Target: b.Workdir, Source: src, Type: mount.TypeBind},
		}
		bconts[i] = cc

		// Mount docker.sock in container if requested.
		if mc.AllowDockerAccess {
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

func readMoldConfig(moldFile string) (*MoldConfig, error) {
	d, err := ioutil.ReadFile(moldFile)
	if err == nil {
		return NewMoldConfig(d)
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
