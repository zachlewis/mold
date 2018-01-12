package main

import (
	"testing"
)

func Test_parseEnvFile(t *testing.T) {
	testMc, _, err := initializeBuild("./testdata/mold.env.yml", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	b := testMc.Build[0]
	b.EnvFiles = append(b.EnvFiles, "i'm not exist")

	_, err = b.GetEnvStrings()
	if err == nil {
		t.Fatalf("should fail with error")
	}

	b.Environment = append(b.Environment, "BAD_VAL")
	_, err = b.GetEnvStrings()
	if err == nil {
		t.Fatalf("should fail with error \"wrong environment value\"")
	}
}
