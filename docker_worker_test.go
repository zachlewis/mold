package main

import (
	"testing"
	"time"
)

/*func TestMain(m *testing.M) {
	testBld, _ = NewDockerWorker(nil)
	code := m.Run()
	testBld.Teardown()
	os.Exit(code)
}*/

func Test_Worker_Configure(t *testing.T) {
	testBc, testBld, _ := initializeBuild("./testdata/mold1.yml", "")
	//testBld, _ := NewDockerWorker(nil)
	if err := testBld.Configure(testBc); err != nil {
		t.Fatal(err)
	}
	if len(testBld.sc) != len(testBc.Services) {
		t.Fatalf("service mismatch have=%d want=%d", len(testBld.sc), len(testBc.Services)-1)
		t.FailNow()
	}
	if len(testBld.bc) != len(testBc.Build) {
		t.Fatal("service mismatch")
	}
	for _, s := range testBld.sc {
		if s.Name == "" {
			t.Fatal("name empty for container", s.Container.Image)
		}
	}

}

func Test_Worker_Build(t *testing.T) {
	//testBld, _ := NewDockerWorker(nil)
	testBc, testBld, _ := initializeBuild("./testdata/mold2.yml", "")
	testBld.Configure(testBc)

	if err := testBld.Setup(); err != nil {
		t.Fatal(err)
	}

	if err := testBld.Build(); err != nil {
		testBld.Teardown()
		t.Fatal(err)
	}

	if err := testBld.Teardown(); err != nil {
		t.Log(err)
		t.Fail()
	}

	for _, v := range testBld.bc {
		t.Log(v.Name, v.Status())
	}
}

func Test_Worker_GeneratesArtifacts(t *testing.T) {
	//testBld, _ := NewDockerWorker(nil)
	//testBld.Configure(testBc)
	testBc, testBld, _ := initializeBuild("./testdata/mold1.yml", "")
	testBld.Configure(testBc)

	if err := testBld.GenerateArtifacts(); err != nil {
		t.Fatal(err)
	}
	if err := testBld.RemoveArtifacts(); err != nil {
		t.Fatal(err)
	}
	if err := testBld.GenerateArtifacts("euforia/mold-test"); err != nil {
		t.Fatal(err)
	}
	testBld.RemoveArtifacts()
	if err := testBld.GenerateArtifacts("foo"); err == nil {
		t.Fatal("should fail with artifact not found")
	}
}

func Test_Worker_GeneratesArtifacts_Abort(t *testing.T) {
	bcfg, bld, _ := initializeBuild("./testdata/mold1.yml", "")
	if err := bld.Configure(bcfg); err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := bld.GenerateArtifacts(); err != nil {
			t.Fatal(err)
		}
	}()

	<-time.After(1500 * time.Millisecond)
	if err := bld.Abort(); err != nil {
		t.Fatal(err)
	}
}

func Test_Worker_Publish_fail(t *testing.T) {
	_, bld, _ := initializeBuild(testBldCfg, "")
	bld.authCfg = nil
	if err := bld.Publish(); err == nil {
		t.Fatal("should fail")
	}
}

func Test_Worker_Publish_fail2(t *testing.T) {
	bcfg, bld, _ := initializeBuild("./testdata/mold4.yml", "")
	if err := bld.Configure(bcfg); err != nil {
		t.Fatal(err)
	}

	if err := bld.Publish(); err == nil {
		t.Fatalf("should fail with image not found: %+v", bcfg.Artifacts.Images)
	}
}

func TestBuildListenOnPort(t *testing.T) {
	bcfg, bld, _ := initializeBuild("./testdata/mold8.yml", "")
	if err := bld.Configure(bcfg); err != nil {
		t.Fatal(err)
	}
	if err := bld.GenerateArtifacts(); err != nil {
		t.Fatal(err)
	}

}
