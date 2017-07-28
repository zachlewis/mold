package main

import "testing"

func Test_gitVersion(t *testing.T) {
	gt, err := newGitVersion(".")
	if err != nil {
		t.Fatal(err)
	}

	if gt.TagVersion() == "0.0.0" {
		t.Fatal("should not be 0.0.0")
	}
	t.Log(gt.Version())
}
