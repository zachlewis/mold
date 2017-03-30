package main

import (
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
