package main

import (
	"testing"
	"time"
)

func Test_LifeCycle(t *testing.T) {
	bc, worker, err := initializeBuild("./testdata/mold3.yml", *dockerURI)
	if err != nil {
		t.Fatal(err)
	}

	//bc, _ := readBuildConfig("./testdata/mold1.yml")
	//bc.RepoName += "-test2"

	//worker, _ := NewDockerWorker(nil)

	lc := NewLifeCycle(worker)
	if err := lc.Run(bc); err != nil {
		t.Fatal(err)
	}
	dw := lc.worker.(*DockerWorker)
	dw.RemoveArtifacts()
}

func Test_LifeCycle_fail(t *testing.T) {
	bc, _ := readBuildConfig("./testdata/mold.fail.yml")
	bc.RepoName += "-test3"

	worker, err := NewDockerWorker(nil)
	if err != nil {
		t.Fatal(err)
	}
	lc := NewLifeCycle(worker)
	if err := lc.Run(bc); err == nil {
		t.Fatal("should fail")
	}
}

func Test_LifeCycle_Abort(t *testing.T) {
	bc, bld, err := initializeBuild("./testdata/mold3.yml", *dockerURI)
	if err != nil {
		t.Fatal(err)
	}

	bc.RepoName += "-test4"

	lc := NewLifeCycle(bld)
	go func() {
		<-time.After(2750 * time.Millisecond)
		if err := lc.Abort(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := lc.Run(bc); err != nil {
		t.Fatal(err)
	}

}

func Test_LifeCycle_RunTarget(t *testing.T) {
	bc, bld, _ := initializeBuild("./testdata/mold1.yml", *dockerURI)

	bc.RepoName += "-test5"

	lc := NewLifeCycle(bld)
	if err := lc.RunTarget(bc, "artifacts", "euforia/mold-test"); err != nil {
		t.Fatal(err)
	}

	if err := lc.RunTarget(bc, "build"); err != nil {
		t.Fatal(err)
	}
}

func Test_LifeCycle_Resolution(t *testing.T) {
	bc, bld, _ := initializeBuild("./testdata/mold7.yml", *dockerURI)
	lc := NewLifeCycle(bld)
	if err := lc.Run(bc); err != nil {
		t.Fatal(err)
	}
}
