package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/docker/docker/pkg/archive"
)

var dockerSockFile = "/var/run/docker.sock"

const (
	dockerIgnoreFile = ".dockerignore"
	linuxDockerURI   = "unix:///var/run/docker.sock"
	windowsDockerURI = "tcp://127.0.0.1:2375"
)

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

// getBuildHash gets sha256 hash of a container config
func getBuildHash(cfg *ContainerConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("Invalid container config -- Empty config")
	}
	name := cfg.Name
	cfg.Name = ""
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("The container config cannot be serialized to json: %s", err)
	}
	cfg.Name = name
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h), nil
}

// returns the name of the image.  it parses out the namespace and tag if provided
func nameFromImageName(imageName string) string {
	iparts := strings.Split(strings.Split(imageName, ":")[0], "/")
	if len(iparts) == 1 {
		return iparts[0]
	}
	return iparts[len(iparts)-1]
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

func getEnvVars(envFile string) ([]string, error) {
	fd, err := os.Open(envFile)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var envs []string
	scanner := bufio.NewScanner(fd)

	for scanner.Scan() {
		envPair := scanner.Text()
		envs = append(envs, envPair)
	}
	return envs, nil
}

func getExcludes(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var excludes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		excludes = append(excludes, scanner.Text())
	}
	return excludes
}

func tarDirectory(srcPath string) (io.ReadCloser, error) {
	excludes := getExcludes(filepath.Join(srcPath, dockerIgnoreFile))
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

func shortContainerName(cName string) string {
	iDigits := strings.LastIndexAny(cName, "-")
	if iDigits > 0 && len(cName) > iDigits+7 {
		cName = cName[:iDigits+7]
	}
	return cName
}

func printUsage() {
	fmt.Printf(`
mold [ options ]

Mold is a tool to perform testing, building, packaging and publishing of
applications completely using Docker.  Application code is tested and built in a
Docker container following by the building for Docker images and publishing to a
registry, all controlled via a single configuration file.

Options:

  -version      Show version

  -app-version  Show the app version from git (default: 0.0.0)

  -init         Initialize a new %s for the project if one does not exist.

  -var          Show value of vairable specified in the configuration file  (default: NA)

  -uri          Docker URI          (default: %s)

  -f            Configuration file  (default: %s)

  -t            Target to build     (default: all)

                build       Only perform the build phase.

                artifacts   Only generate artifacts.  Specific artifacts can be built
                            using artifacts/<image_name> as the target where
                            <image_name> would be that as specified in your
                            configuration.

                publish     Only publish artifacts.  Specific artifacts can be published
                            using publish/<image_name> as the target where
                            <image_name> would be that as specified in your
                            configuration.

`, defaultBuildConfigName, *dockerURI, *buildFile)
}
