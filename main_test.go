package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func Test_printVersion(t *testing.T) {
	printVersion()
	printUsage()
}

func Test_MoldInit(t *testing.T) {
	d, err := ioutil.TempDir("", "mold")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(d)

	if err := initializeMoldConfig(d); err != nil {
		//t.Error("should fail to init .mold.yml")
		t.Fatal(err)
	}
}

func Test_getVar_should_pass(t *testing.T) {
	val, err := getVar("var1", "testdata/mold5.yml")
	if err == nil {
		fmt.Printf("%s\n", val)
		if val != "val1" {
			t.Fatal("got incorrect value")
		}
	} else {
		t.Fatal(err)
	}
}

func Test_getVar_should_fail(t *testing.T) {
	_, err := getVar("var0", "testdata/mold5.yml")
	if err == nil {
		t.Fatal("should have failed because the variable is not specified in the config file")
	}
}

func Test_getVar_should_fail2(t *testing.T) {
	_, err := getVar("var1", "testdata/mold1.yml")
	if err == nil {
		t.Fatal("should have failed because the variable is not specified in the config file")
	}
}

func Test_getVar_should_fail3(t *testing.T) {
	_, err := getVar("var0", "testdata/mold6.yml")
	if err == nil {
		t.Fatal("should have failed because the bad config format")
	}
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
