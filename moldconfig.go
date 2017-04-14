package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const defaultBuildConfigName = ".mold.yml"

// MoldConfig holds the complete build configuration
type MoldConfig struct {
	// Build name.  This is the name with the branch and commit included and calculated internally.
	//name string

	// Name of the repo.  This is different from the above name in that it is just the name of the project
	RepoName string `yaml:"-"`
	// Git url
	RepoURL string `yaml:"-"`
	// Tag or branch to build
	BranchTag string `yaml:"-"`
	// LastCommit for the branch
	LastCommit string `yaml:"-"`
	// Context is the root of the build.  This defaults to the current working
	// directory.
	Context string `yaml:",omitempty"`
	// Service i.e. containers needed to perform build
	Services []DockerRunConfig
	// Builds to perform
	Build []DockerRunConfig
	// Docker images to generate
	Artifacts Artifacts
	// Allow docker daemon access in the container
	AllowDockerAccess bool `yaml:"docker,omitempty"`

	Variables map[string]string `yaml:",omitempty"`
	// stores version information from git
	gitVersion *gitVersion
}

func DefaultMoldConfig(name string) *MoldConfig {
	return &MoldConfig{
		Build: []DockerRunConfig{DockerRunConfig{}},
		Artifacts: Artifacts{
			Images:  []ImageConfig{ImageConfig{Name: name}},
			Publish: []string{"master"},
		},
	}
}

// NewMoldConfig creates a new config from yaml formatted bytes
func NewMoldConfig(fileBytes []byte) (*MoldConfig, error) {
	var mc MoldConfig
	err := yaml.Unmarshal(fileBytes, &mc)
	if err != nil {
		return nil, err
	}

	mc.gitVersion, _ = newGitVersion(".")

	// Set current working directory if not specified
	if mc.Context == "" || mc.Context == "." || mc.Context == "./" {
		if mc.Context, err = os.Getwd(); err != nil {
			return nil, err
		}
	}

	for i, v := range mc.Build {
		if v.Shell == "" {
			mc.Build[i].Shell = "/bin/sh"
		}
	}
	mc.setBuildEnvVars()

	// Set artifact defaults
	for i, v := range mc.Artifacts.Images {
		// set the context to the working dir if not supplied
		if len(v.Context) == 0 {
			mc.Artifacts.Images[i].Context = mc.Context
		}
	}
	mc.Artifacts.setDefaults()
	mc.normalizeArtifactsImageTags()

	if err = mc.Artifacts.ValidateImageConfigs(); err != nil {
		return nil, err
	}

	mc.checkRepoInfo()
	mc.readEnvVars()

	// try to set the name based on the repo url.
	if mc.RepoURL != "" {
		if pp := strings.Split(mc.RepoURL, "/"); len(pp) > 1 {
			if n := strings.TrimSuffix(pp[len(pp)-1], ".git"); n != "" {
				mc.RepoName = n
			}
		}
	}

	return &mc, err
}

// Normalize image tag vars.

func (mc *MoldConfig) normalizeArtifactsImageTags() {
	for i := range mc.Artifacts.Images {
		mc.Artifacts.Images[i].ReplaceTagVars("${APP_VERSION}", mc.gitVersion.Version())
		mc.Artifacts.Images[i].ReplaceTagVars("${APP_VERSION_SHORT}", mc.gitVersion.TagVersion())
		mc.Artifacts.Images[i].ReplaceTagVars("${APP_COMMIT}", mc.gitVersion.Commit())
		mc.Artifacts.Images[i].ReplaceTagVars("${APP_COMMIT_INDEX}", fmt.Sprintf("%d", mc.gitVersion.distance))
	}
}

// Name returns the name of the build image to create
func (mc *MoldConfig) Name() string {
	if len(mc.LastCommit) > 7 {
		return mc.RepoName + "-" + mc.BranchTag + "-" + mc.LastCommit[:8]
	} else if len(mc.BranchTag) > 0 {
		return mc.RepoName + "-" + mc.BranchTag
	}
	return mc.RepoName
}

// Inject app version as env. var. to build container
func (mc *MoldConfig) setBuildEnvVars() {
	evars := []string{
		"APP_VERSION=" + mc.gitVersion.Version(),
		"APP_VERSION_SHORT=" + mc.gitVersion.TagVersion(),
		"APP_COMMIT=" + mc.gitVersion.Commit(),
		fmt.Sprintf("APP_COMMIT_INDEX=%d", mc.gitVersion.distance),
	}

	for i, v := range mc.Build {
		if v.Environment == nil {
			mc.Build[i].Environment = evars
		} else {
			mc.Build[i].Environment = append(mc.Build[i].Environment, evars...)
		}
	}
}

// check and set repo info and naming structure - RE-VISIT
func (mc *MoldConfig) checkRepoInfo() {

	name, bt, lc := getRepoInfo(mc.Context)
	if len(mc.RepoName) == 0 && len(name) > 0 {
		mc.RepoName = name
	}
	if len(mc.BranchTag) == 0 && len(bt) > 0 {
		mc.BranchTag = bt
	}
	if len(mc.LastCommit) == 0 && len(lc) > 0 {
		mc.LastCommit = lc
	}
}

// read env vars and config.  These take precedence over all configs overriding
// anything prior - RE-VISIT
func (mc *MoldConfig) readEnvVars() {
	if cmt := os.Getenv("GIT_COMMIT"); len(cmt) > 7 {
		mc.LastCommit = cmt[:8]
	}
	if rurl := os.Getenv("GIT_URL"); len(rurl) > 0 {
		mc.RepoURL = rurl
	}
	if branchTag := os.Getenv("GIT_BRANCH"); len(branchTag) > 0 {
		pp := strings.Split(branchTag, "/")
		if len(pp) > 0 {
			mc.BranchTag = pp[len(pp)-1]
		}
	}
}
