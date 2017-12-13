package main

import "testing"

func Test_Artifacts(t *testing.T) {
	mc, err := readMoldConfig("./testdata/mold2.yml")
	if err != nil {
		t.Fatal(err)
	}
	mc.RepoName += "-test4"
	if len(mc.Artifacts.Publish) < 1 {
		t.Fatal("publish should be non-zero")
	}

	registry := "test-registry"
	mc.Artifacts.Registry = registry
	mc.Artifacts.setDefaults()
	for _, img := range mc.Artifacts.Images {
		if len(img.Registry) == 0 && img.Registry != registry {
			t.Fatalf("image registry: %s; want: %s", img.Registry, registry)
		}
	}
}
