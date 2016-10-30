package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

var (
	testBldCfg = "testdata/mold1.yml"
	testBc     *BuildConfig
	testBld    *DockerWorker
)

func TestMain(m *testing.M) {
	testBld, _ = NewDockerWorker(nil)
	code := m.Run()
	testBld.Teardown()
	os.Exit(code)
}

func Test_NewBuildConfig(t *testing.T) {
	b, err := ioutil.ReadFile(testBldCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if testBc, err = NewBuildConfig(b); err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	testBc.Name += "-test1"
	b, _ = json.MarshalIndent(testBc, "", "  ")

	t.Logf("%s\n", b)

	for _, v := range testBc.Artifacts.Images {
		if v.Dockerfile == "" {
			t.Fatal("docker file empty")
		}
	}
	bimg, err := testBc.Artifacts.Images[0].BaseImage()
	if err != nil {
		t.Fatal(err)
	}
	if bimg != "alpine" {
		t.Fatal("base image should be alpine")
	}

	if _, err = NewBuildConfig(b[1:]); err == nil {
		t.Fatal("should fail")
	}
}
