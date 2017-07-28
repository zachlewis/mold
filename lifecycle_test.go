package main

import (
	"testing"
	"time"
)

func Test_LifeCycle_Run(t *testing.T) {
	mc, worker, err := initializeBuild("./testdata/mold3.yml", *dockerURI)
	if err != nil {
		t.Fatal(err)
	}

	lc := NewLifeCycle(worker)
	if err := lc.Run(mc); err != nil {
		t.Fatal(err)
	}
	dw := lc.worker.(*DockerWorker)
	dw.RemoveArtifacts()
}

func Test_LifeCycle_buildless(t *testing.T) {
	mc, worker, err := initializeBuild("./testdata/mold.buildless.yml", *dockerURI)
	if err != nil {
		t.Fatal(err)
	}

	lc := NewLifeCycle(worker)
	if err := lc.Run(mc); err != nil {
		t.Fatal(err)
	}
	dw := lc.worker.(*DockerWorker)
	dw.RemoveArtifacts()
}

func Test_LifeCycle_fail(t *testing.T) {
	mc, _ := readMoldConfig("./testdata/mold.fail.yml")
	mc.RepoName += "-test3"

	worker, err := NewDockerWorker(nil)
	if err != nil {
		t.Fatal(err)
	}
	lc := NewLifeCycle(worker)
	if err := lc.Run(mc); err == nil {
		t.Fatal("should fail")
	}
}

func Test_LifeCycle_Abort(t *testing.T) {
	mc, worker, err := initializeBuild("./testdata/mold3.yml", *dockerURI)
	if err != nil {
		t.Fatal(err)
	}

	mc.RepoName += "-test4"

	lc := NewLifeCycle(worker)
	go func() {
		<-time.After(2750 * time.Millisecond)
		if err := lc.Abort(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := lc.Run(mc); err != nil {
		t.Fatal(err)
	}

}

func Test_LifeCycle_RunTarget(t *testing.T) {
	mc, worker, _ := initializeBuild("./testdata/mold1.yml", *dockerURI)

	mc.RepoName += "-test5"

	lc := NewLifeCycle(worker)
	if err := lc.RunTarget(mc, "artifacts", "d3sw/mold-test"); err != nil {
		t.Fatal(err)
	}

	if err := lc.RunTarget(mc, "build"); err != nil {
		t.Fatal(err)
	}
}

func Test_LifeCycle_Resolution(t *testing.T) {
	mc, worker, _ := initializeBuild("./testdata/mold7.yml", *dockerURI)
	lc := NewLifeCycle(worker)
	if err := lc.Run(mc); err != nil {
		t.Fatal(err)
	}
}

func Test_LifeCycle_multi_artifact(t *testing.T) {
	mc, worker, _ := initializeBuild("./testdata/mold.multi-art.yml", *dockerURI)
	lc := NewLifeCycle(worker)
	if err := lc.Run(mc); err != nil {
		t.Fatal(err)
	}
}
