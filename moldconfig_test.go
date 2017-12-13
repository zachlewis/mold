package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

var (
	testMoldCfg = "testdata/mold1.yml"
)

func Test_DefaultMoldConfig(t *testing.T) {
	mc := DefaultMoldConfig("test")
	b, e := yaml.Marshal(mc)
	if e != nil {
		t.Fatal(e)
	}

	fmt.Printf("%s", b)
}

func Test_NewMoldConfig(t *testing.T) {
	b, err := ioutil.ReadFile(testMoldCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if _, ok := os.LookupEnv("GIT_URL"); !ok {
		os.Setenv("GIT_URL", "https://github.com/dummy/dummy.git")
		defer os.Unsetenv("GIT_URL")
	}

	if _, ok := os.LookupEnv("GIT_BRANCH"); !ok {
		os.Setenv("GIT_BRANCH", "origin/master")
		defer os.Unsetenv("GIT_BRANCH")
	}

	if _, ok := os.LookupEnv("GIT_COMMIT"); !ok {
		os.Setenv("GIT_COMMIT", "1234567890123456")
		defer os.Unsetenv("GIT_COMMIT")
	}

	testMc, err := NewMoldConfig(b)
	if err != nil {
		t.Fatal(err)
	}

	if len(testMc.LastCommit) == 0 {
		t.Log("last commit should be set")
		t.Fail()
	}

	if len(testMc.Name()) == 0 {
		t.Log("name should be set")
		t.Fail()
	}

	if len(testMc.Name()) == 0 {
		t.Log("name should be set")
		t.Fail()
	}
	for _, v := range testMc.Build {
		if v.Image == "" {
			t.Fatal("image should be set")

		}
	}

	if !strings.HasPrefix(testMc.Context, "/") {
		t.Error("context path not *nix")
	}

	testMc.RepoName += "-test1"
	b, _ = json.MarshalIndent(testMc, "", "  ")
	t.Logf("%s\n", b)
	t.Log(testMc.Name())

	for _, v := range testMc.Artifacts.Images {
		if v.Dockerfile == "" {
			t.Fatal("docker file empty")
		}
	}
	bimg, err := testMc.Artifacts.Images[0].BaseImage()
	if err != nil {
		t.Fatal(err)
	}
	if bimg != "alpine" {
		t.Fatal("base image should be alpine")
	}

	if _, err = NewMoldConfig(b[1:]); err == nil {
		t.Fatal("should fail")
	}
}

func TestReadBuildPort(t *testing.T) {
	buildFile := "testdata/mold1.yml"

	b, err := ioutil.ReadFile(buildFile)
	if err != nil {
		t.Errorf("Could not read %v", buildFile)
	}

	testBc, err := NewMoldConfig(b)
	if err != nil {
		t.Errorf("Could build config from file %v", buildFile)
	}

	if len(testBc.Build) == 0 {
		t.Errorf("Expected build: section of %v but didn't find it", buildFile)
	}

	if len(testBc.Build[0].Ports) == 0 {
		t.Errorf("Expected ports: section of %v but didn't find it", buildFile)
	}

	if testBc.Build[0].Ports[0] != "61868:61868" {
		t.Errorf("Expected 61868:61868 in ports section of %v but didn't find it", buildFile)
	}
}

func Test_NewMoldConfig_ImageWithoutName(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/mold.image.noname.yml")

	if err != nil {
		t.Fatal(err)
	}

	_, err = NewMoldConfig([]byte(b))

	if err == nil {
		t.Error("Expected error to be thrown")
	}

	expected := "image without a name is not allowed"
	if err.Error() != expected {
		t.Errorf("Expected '%s' error message, but got '%s'", expected, err)
	}
}
