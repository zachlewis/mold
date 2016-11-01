package main

import (
	"os"
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
	// Working dir of the whole build.  This is essentially the root of the code
	// repo on the host.
	WorkingDir string
	// Service i.e. containers needed to perform build
	Services []DockerRunConfig
	// Builds to perform
	Build []DockerRunConfig
	// Docker images to generate
	Artifacts Artifacts
	//Artifacts []ImageConfig

	// Notifications through out the build process
	Notifications MultiNotification

	AllowDockerAccess bool `yaml:"docker"` // Mount docker socket to the container.
}

// NewBuildConfig creates a new config from yaml formatted bytes
func NewBuildConfig(fileBytes []byte) (*BuildConfig, error) {
	var bc BuildConfig
	err := yaml.Unmarshal(fileBytes, &bc)
	if err != nil {
		return nil, err
	}
	// Set working directory
	if bc.WorkingDir, err = os.Getwd(); err == nil {
		// Set artifact defaults
		for i, v := range bc.Artifacts.Images {
			// set the context to the working dir if not supplied
			if len(v.Context) == 0 {
				bc.Artifacts.Images[i].Context = bc.WorkingDir
			}
		}
		bc.Artifacts.setDefaults()
		bc.readEnvVars()
		//bc.findRepoInfo()
	}

	return &bc, err
}

// extract name from url
func (bc *BuildConfig) setNameFromRepoURL() {
	pp := strings.Split(bc.RepoURL, "/")
	if len(pp) > 0 {
		bc.Name = strings.TrimSuffix(pp[len(pp)-1], ".git")
	}
}

func (bc *BuildConfig) readEnvVars() {

	bc.LastCommit = os.Getenv("GIT_COMMIT")
	if len(bc.RepoURL) == 0 {
		bc.RepoURL = os.Getenv("GIT_URL")
	}
	if len(bc.BranchTag) == 0 {
		if gb := os.Getenv("GIT_BRANCH"); len(gb) > 0 {
			pp := strings.Split(gb, "/")
			if len(pp) > 0 {
				bc.BranchTag = pp[len(pp)-1]
			}
		}
	}
	if len(bc.Name) == 0 {
		bc.setNameFromRepoURL()
	}
	// set unique name based on name branch and commit
	bc.Name += "-" + bc.BranchTag
	if len(bc.LastCommit) > 7 {
		bc.Name += "-" + bc.LastCommit[:8]
	}
}

/*
// try to get name and branch info from the working dir.
func (bc *BuildConfig) findRepoInfo() {
	bc.Name = filepath.Base(bc.WorkingDir)

	if b, err := ioutil.ReadFile(filepath.Join(bc.WorkingDir, ".git/HEAD")); err == nil {
		line := string(bytes.TrimSuffix(b, []byte("\n")))
		ref := strings.Split(line, " ")[1]
		pp := strings.Split(ref, "/")
		bc.BranchTag = pp[len(pp)-1]

		if cmt, err := ioutil.ReadFile(filepath.Join(bc.WorkingDir, ".git", ref)); err == nil {
			bc.LastCommit = string(cmt[:8])
			bc.Name += "-" + bc.BranchTag + "-" + bc.LastCommit
		}
	}

}
*/
