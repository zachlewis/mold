package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

const dockerSockFile = "/var/run/docker.sock"

func printVersion() {
	fmt.Printf("%s branch=%s commit=%s buildtime=%s\n", VERSION, branch, commit, buildtime)
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
		bcs[i] = &ContainerConfig{
			Container: &container.Config{
				Image: b.Image,
				Cmd:   b.Commands,
				Env:   b.Environment,
			},
			Network: &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{},
			},
			Host: &container.HostConfig{},
		}
	}
	return bcs
}

// assembleBuildContainers assembles container configs from user provided build config
func assembleBuildContainers(bc *BuildConfig) []*ContainerConfig {
	bconts := make([]*ContainerConfig, len(bc.Build))
	for i, b := range bc.Build {
		bconts[i] = &ContainerConfig{
			Container: &container.Config{
				Image:      b.Image,
				WorkingDir: b.Workdir,
				Volumes: map[string]struct{}{
					b.Workdir: struct{}{},
				},
				Cmd: []string{"/bin/bash", "-cex", b.BuildCmds()},
				Env: b.Environment,
				//Hostname:    name,
				//Env:         svc.Environment,
				//Labels:      nil,
				//Healthcheck: nil,
			},
			Network: &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{},
			},
			Host: &container.HostConfig{
				Binds: []string{fmt.Sprintf("%s:%s", bc.WorkingDir, b.Workdir)},
				Mounts: []mount.Mount{
					mount.Mount{Target: b.Workdir, Source: bc.WorkingDir, Type: mount.TypeBind},
				},
				//DNS:        []string{},
				//DNSOptions: []string{},
				//DNSSearch:  []string{},
			},
		}
		// Mount docker.sock in container if requested.
		if bc.AllowDockerAccess {
			bconts[i].Container.Volumes[dockerSockFile] = struct{}{}
			bconts[i].Host.Binds = append(bconts[i].Host.Binds, fmt.Sprintf("%s:%s", dockerSockFile, dockerSockFile))
			bconts[i].Host.Mounts = append(bconts[i].Host.Mounts,
				mount.Mount{Target: dockerSockFile, Source: dockerSockFile, Type: mount.TypeBind})
		}
	}
	return bconts
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

// Tar a given directory and return the underlying reader
func tarDirectory(dir string, wr io.ReadWriter) (io.Reader, error) {
	// Use provided writer if supplied
	w := wr
	if w == nil {
		w = new(bytes.Buffer)
	}

	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var link string
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(path); err != nil {
				return err
			}
		}

		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}

		// write relative path
		relPath := strings.TrimPrefix(path, dir)
		header.Name = relPath
		if err = tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() { //nothing more to do for non-regular
			return nil
		}

		fh, err := os.Open(path)
		if err == nil {
			defer fh.Close()
			_, err = io.Copy(tw, fh)
		}
		return err
	})

	return w, err
}
