package main

import "testing"

func Test_gitVersion(t *testing.T) {
	gt, err := newGitVersion(".")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(gt.Version())
}

/*func Test_getGitVersion(t *testing.T) {
	abs, _ := filepath.Abs(".")
	a := getGitVersion(abs)
	if a.Version() == "0.0.0" {
		t.Error("version should not be default")
	}

	if a.index == 0 {
		t.Error("count should be greater than 0")
	}
	if a.head.Hash().IsZero() {
		t.Error("hash should not be zero")
	}
}

func Test_getGitVersion2(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "ggi-test")
	defer os.RemoveAll(tmp)

	a := getGitVersion(tmp)
	if a.Version() != "0.0.0" {
		t.Error("version should be 0.0.0")
	}
	if a.index != 0 {
		t.Error("cnt should be 0")
	}
	if a.head != nil {
		t.Error("head should be nil")
	}
}*/
