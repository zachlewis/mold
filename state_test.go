package main

import (
	"testing"
)

func Test_ToString(t *testing.T) {
	c := (*cache)(nil)
	if c.ToString() != "" {
		t.Fatal("should be empty string")
	}
	c = &cache{
		Name: "n1",
		Tag:  "t1",
	}

	if c.ToString() != "n1:t1" {
		t.Fatal("bad string")
	}
}
