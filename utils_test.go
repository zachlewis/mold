package main

import (
	"path/filepath"
	"testing"
)

func equal(a, b []string) bool {
	if a == nil || b == nil || len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Test_getExcludes(t *testing.T) {
	expectedExcludes := []string{"huge_file", "important_dir_with_huge_files/*", "dir_with_huge_files"}
	excludes := getExcludes(filepath.Join("testdata/", dockerIgnoreFile))
	if !equal(excludes, expectedExcludes) {
		t.Fatal("should return ", expectedExcludes, "; got ", excludes)
	}
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

func Test_toDockerWinPath(t *testing.T) {
	unixDockerPath := "test_pr/test_path"
	expectedPath := "//test_pr/test_path"
	if s := toDockerWinPath(unixDockerPath); s != expectedPath {
		t.Fatalf("expected %s, got %s\n", expectedPath, s)
	}
}
