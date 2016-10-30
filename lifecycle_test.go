package main

import "testing"

func Test_LifeCycle(t *testing.T) {
	bc, _ := readBuildConfig("./testdata/mold1.yml")
	bc.Name += "-test2"

	lc := NewLifeCycle(testBld)
	if err := lc.Run(bc); err != nil {
		t.Fatal(err)
	}
	dw := lc.worker.(*DockerWorker)
	dw.RemoveArtifacts()
}

func Test_LifeCycle_fail(t *testing.T) {
	bc, _ := readBuildConfig("./testdata/mold2.yml")
	bc.Name += "-test3"

	lc := NewLifeCycle(testBld)
	if err := lc.Run(bc); err == nil {
		t.Fatal("should fail")
	}
}
