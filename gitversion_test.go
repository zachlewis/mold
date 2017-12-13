package main

import (
	"testing"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

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

func Test_gitTag(t *testing.T) {
	gt, err := newGitVersion(".")
	if err != nil {
		t.Fatal(err)
	}

	if gt.getTag("xxx") != "" {
		t.Fatal("should be empty string")
	}

	h := plumbing.NewHash("xxx")
	gt.tags[h] = plumbing.NewReferenceFromStrings("xxx", "yyy")
	if gt.getTag("xxx") == "" {
		t.Fatal("tag should be found")
	}
}
