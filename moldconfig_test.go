package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"
)

var (
	testMoldCfg     = "testdata/mold1.yml"
	testMoldfileWin = "testdata/mold.win.yml"
)

func Test_NewMoldConfig(t *testing.T) {
	b, err := ioutil.ReadFile(testMoldCfg)
	if err != nil {
		t.Fatalf("%s", err)
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

	if testBc.Build[0].Ports[0] != "5432:5432" {
		t.Errorf("Expected 5432:5432 in ports section of %v but didn't find it", buildFile)
	}
}
