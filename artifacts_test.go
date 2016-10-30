package main

import "testing"

func Test_Artifacts(t *testing.T) {
	bc, err := readBuildConfig("./testdata/mold2.yml")
	if err != nil {
		t.Fatal(err)
	}
	bc.Name += "-test4"
	if len(bc.Artifacts.Publish) < 1 {
		t.Fatal("publish should be non-zero")
	}
}
