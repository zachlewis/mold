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
