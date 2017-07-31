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

func Test_getBuildHash(t *testing.T) {
	moldFile := "testdata/mold9.yml"
	cfg, err := readMoldConfig(moldFile)
	if err != nil {
		t.Fatal(err)
	}
	bc, err := assembleBuildContainers(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var hs []string
	for _, cc := range bc {
		h, err := getBuildHash(cc)
		if err != nil {
			t.Fatal(err)
		}
		hs = append(hs, h)
	}
	if hs[0] != hs[1] {
		t.Fatalf("Same hash should be generated for identical config")
	}
	if hs[1] == hs[2] {
		t.Fatalf("Different hash should be generated for different config")
	}
	if hs[2] == hs[3] {
		t.Fatalf("Different hash should be generated for different config")
	}
}
