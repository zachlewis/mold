package main

import (
	"strings"
	"testing"
)

func Test_ImageConfig(t *testing.T) {
	ic := ImageConfig{
		Name:     "name",
		Tags:     []string{"${REPLACE}-rc1", "${REPLACE}"},
		Registry: "registry",
	}

	if err := ic.Validate(); err != nil {
		t.Fatal(err)
	}

	ic.ReplaceTagVars("${REPLACE}", "1.1.1")
	for _, tag := range ic.Tags {
		if strings.Contains(tag, "${REPLACE}") {
			t.Error("var not replaced")
		}
	}

	out := ic.DefaultRegistryPaths()
	if len(out) != 3 {
		t.Fatal("invalid registry count")
	}
	t.Log(out)

	out = ic.CustomRegistryPaths()
	if len(out) != 3 {
		t.Fatal("invalid registry count")
	}
	t.Log(out)
}

func Test_RegistryPaths(t *testing.T) {
	i := ImageConfig{}
	if paths := i.RegistryPaths(); paths == nil {
		t.Fatalf("should be default registry path")
	}
	if paths := i.RegistryPaths(); paths == nil {
		t.Fatalf("should be custom registry path")
	}
}
