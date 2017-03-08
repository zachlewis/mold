package main

import (
	"os"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
)

const defaultBuildConfigName = ".mold.yml"

// BuildConfig holds the complete build configuration
type BuildConfig struct {
	// Project name
	Name string
	// Git url
	RepoURL string
	// Tag or branch to build
	BranchTag string
	// LastCommit for the branch
	LastCommit string
	// Context is the root of the build.  This defaults to the current working
	// directory.
	Context string
	// Service i.e. containers needed to perform build
	Services []DockerRunConfig
	// Builds to perform
	Build []DockerRunConfig
	// Docker images to generate
	Artifacts Artifacts
	// Notifications through out the build process
	Notifications MultiNotification
	// Allow docker daemon access in the container
	AllowDockerAccess bool `yaml:"docker"`

	Variables map[string]string
}

// NewBuildConfig creates a new config from yaml formatted bytes
func NewBuildConfig(fileBytes []byte) (*BuildConfig, error) {
	var bc BuildConfig
	err := yaml.Unmarshal(fileBytes, &bc)
	if err != nil {
		return nil, err
	}

	// Set current working directory if not specified
	if bc.Context == "" || bc.Context == "." || bc.Context == "./" {
		if bc.Context, err = os.Getwd(); err != nil {
			return nil, err
		}
	}
	// Account for windows unc paths
	if runtime.GOOS == "windows" {
		p1 := strings.Replace(bc.Context, `\`, "/", -1)
		for i, c := range p1 {
			if c == '/' {
				bc.Context = p1[i:]
				break
			}
		}
	}

	// Set artifact defaults
	for i, v := range bc.Artifacts.Images {
		// set the context to the working dir if not supplied
		if len(v.Context) == 0 {
			bc.Artifacts.Images[i].Context = bc.Context
		}
	}
	bc.Artifacts.setDefaults()
	bc.checkRepoInfo()
	bc.readEnvVars()

	// try to set the name based on the repo url.
	if bc.RepoURL != "" {
		if pp := strings.Split(bc.RepoURL, "/"); len(pp) > 1 {
			if n := strings.TrimSuffix(pp[len(pp)-1], ".git"); n != "" {
				bc.Name = n
			}
		}
	}

	// set unique name based on name, branch and commit
	bc.Name += "-" + bc.BranchTag
	if len(bc.LastCommit) > 7 {
		bc.Name += "-" + bc.LastCommit[:8]
	}

	return &bc, err
}

// check and set repo info and naming structure
func (bc *BuildConfig) checkRepoInfo() {

	name, bt, lc := getRepoInfo(bc.Context)
	if len(bc.Name) == 0 && len(name) > 0 {
		bc.Name = name
	}
	if len(bc.BranchTag) == 0 && len(bt) > 0 {
		bc.BranchTag = bt
	}
	if len(bc.LastCommit) == 0 && len(lc) > 0 {
		bc.LastCommit = lc
	}
}

// read env vars and config.  These take precedence over all configs overriding
// anything prior
func (bc *BuildConfig) readEnvVars() {
	if cmt := os.Getenv("GIT_COMMIT"); len(cmt) > 7 {
		bc.LastCommit = cmt[:8]
	}
	if rurl := os.Getenv("GIT_URL"); len(rurl) > 0 {
		bc.RepoURL = rurl
	}
	if branchTag := os.Getenv("GIT_BRANCH"); len(branchTag) > 0 {
		pp := strings.Split(branchTag, "/")
		if len(pp) > 0 {
			bc.BranchTag = pp[len(pp)-1]
		}
	}
}
