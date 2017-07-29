package main

import (
	"os"
	"testing"
	"time"
)

func Test_Worker_Configure(t *testing.T) {
	testMc, worker, _ := initializeBuild("./testdata/mold1.yml", "")
	if err := worker.Configure(testMc); err != nil {
		t.Fatal(err)
	}
	if len(worker.serviceStates) != len(testMc.Services) {
		t.Fatalf("service mismatch have=%d want=%d", len(worker.serviceStates), len(testMc.Services)-1)
		t.FailNow()
	}
	if len(worker.buildStates) != len(testMc.Build) {
		t.Fatal("service mismatch")
	}
	for _, s := range worker.serviceStates {
		if s.Name == "" {
			t.Fatal("name empty for container", s.Container.Image)
		}
	}

}
func Test_Worker_Configure_ImgCache(t *testing.T) {
	testMc, worker, _ := initializeBuild("./testdata/mold9.yml", "")
	if err := worker.Configure(testMc); err != nil {
		t.Fatal(err)
	}
	if len(worker.buildStates) != len(testMc.Build) {
		t.Fatal("service mismatch")
	}
	if worker.buildStates[0].ImgCache.Tag != "24739ed1917361a71da27424c1fec6319ff35d25d8031e210b7fcc3cee84b874" {
		t.Fatalf("ImgCache tag value is incorrect: %s", worker.buildStates[0].ImgCache.Tag)
	}
	if worker.buildStates[1].ImgCache != nil {
		t.Fatalf("ImgCache should be nil when not specified in the mold config")
	}
}

func Test_Worker_Build(t *testing.T) {
	testMc, worker, _ := initializeBuild("./testdata/mold2.yml", "")
	worker.Configure(testMc)

	if err := worker.Setup(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Build(); err != nil {
		worker.Teardown()
		t.Fatal(err)
	}

	if err := worker.Teardown(); err != nil {
		t.Log(err)
		t.Fail()
	}

	for _, v := range worker.buildStates {
		t.Log(v.Name, v.Status())
	}
}

func TestAppendOsEnv_onlyReplacesWhenNoEqualsFound(t *testing.T) {
	drc := DockerRunConfig{
		Environment: []string{"DONT_REPLACE=same", "REPLACE"},
	}
	if err := os.Setenv("DONT_REPLACE", "different"); err != nil {
		t.Fatalf("Failed to set up test with: %v", err)
	}
	if err := os.Setenv("REPLACE", "different"); err != nil {
		t.Fatalf("Failed to set up test with: %v", err)
	}

	resultEnv, _ := appendOsEnv(drc.Environment)

	if resultEnv[0] != "DONT_REPLACE=same" {
		t.Errorf("Should not have replaced environment with '=' but did. Got %v", resultEnv[0])
	}
	if resultEnv[1] != "REPLACE=different" {
		t.Errorf("Did not append environment value. Got %v ", resultEnv[1])
	}
}

func TestAppendOsEnv_allowsEmptyEnvVar(t *testing.T) {
	input := []string{"VAL="}
	if err := os.Setenv("VAL", "shoudntReplaceThis"); err != nil {
		t.Fatalf("Failed to set up test with: %v", err)
	}

	resultEnv, _ := appendOsEnv(input)

	if resultEnv[0] != "VAL=" {
		t.Errorf("Should allow empty var declarations")
	}
}

func TestAppendOsEnv_wantedButNotProvidedError(t *testing.T) {
	drc := DockerRunConfig{
		Environment: []string{"WANTED_BUT_NOT_PROVIDED"},
	}

	_, err := appendOsEnv(drc.Environment)

	if err == nil {
		t.Error("Expected an error for a value specified but not provided")
	}
}

func Test_Worker_GeneratesArtifacts(t *testing.T) {
	testMc, worker, _ := initializeBuild("./testdata/mold1.yml", "")
	worker.Configure(testMc)

	if err := worker.GenerateArtifacts(); err != nil {
		t.Fatal(err)
	}
	if err := worker.RemoveArtifacts(); err != nil {
		t.Fatal(err)
	}
	if err := worker.GenerateArtifacts("d3sw/mold-test"); err != nil {
		t.Fatal(err)
	}
	worker.RemoveArtifacts()
	if err := worker.GenerateArtifacts("foo"); err == nil {
		t.Fatal("should fail with artifact not found")
	}
}

func Test_Worker_GeneratesArtifacts_Abort(t *testing.T) {
	testMc, worker, _ := initializeBuild("./testdata/mold1.yml", "")
	if err := worker.Configure(testMc); err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := worker.GenerateArtifacts(); err != nil {
			t.Fatal(err)
		}
	}()

	<-time.After(1500 * time.Millisecond)
	if err := worker.Abort(); err != nil {
		t.Fatal(err)
	}
}

func Test_Worker_Publish_fail(t *testing.T) {
	_, worker, _ := initializeBuild(testMoldCfg, "")
	worker.authCfg = nil
	if err := worker.Publish(); err == nil {
		t.Fatal("should fail")
	}
}

func Test_Worker_Publish_fail2(t *testing.T) {
	testMc, worker, _ := initializeBuild("./testdata/mold4.yml", "")
	if err := worker.Configure(testMc); err != nil {
		t.Fatal(err)
	}

	if err := worker.Publish(); err == nil {
		t.Fatalf("should fail with image not found: %+v", testMc.Artifacts.Images)
	}
}

func TestBuildListenOnPort(t *testing.T) {
	testMc, worker, _ := initializeBuild("./testdata/mold8.yml", "")
	if err := worker.Configure(testMc); err != nil {
		t.Fatal(err)
	}
	if err := worker.GenerateArtifacts(); err != nil {
		t.Fatal(err)
	}

}
