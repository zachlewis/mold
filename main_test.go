package main

import "testing"

func Test_printVersion(t *testing.T) {
	printVersion()
	printUsage()
}

func Test_parseTarget(t *testing.T) {
	p, s := parseTarget("build/foo/bar")
	if p == "" || s == "" {
		t.Fatal("failed to parse target")
	}

	p, s = parseTarget("foo")
	if p == "" || s != "" {
		t.Fatal("failed to parse target")
	}

	p, s = parseTarget("")
	if p != "" || s != "" {
		t.Fatal("failed to parse target")
	}
}
