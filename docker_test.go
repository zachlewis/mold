package main

import (
	"testing"
)

func Test_Docker(t *testing.T) {
	d, err := NewDocker("")
	if err != nil {
		t.Fatal(err)
	}
	if d.Client() == nil {
		t.Fatal("failed to init docker client")
	}
	d = nil

	if d, err = NewDocker("unix:///var/run/docker.sock"); err != nil {
		t.Fatal(err)
	}
	if d.Client() == nil {
		t.Fatal("failed to init docker client")
	}
}
