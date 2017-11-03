package main

import (
	"os"
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

func Test_Docker_PullImage(t *testing.T) {
	d, _ := NewDocker("")
	if err := d.PullImage("busybox:latest", nil, os.Stdout, ""); err != nil {
		t.Fatal(err)
	}
	d.RemoveImage("busybox:latest", true, false)

	if err := d.PullImage("nosuchrepo:latest", nil, os.Stdout, ""); err == nil {
		t.Fatal("Error should be reported when failing to pull an image")
	}
}
