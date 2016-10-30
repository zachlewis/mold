package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	dockerURI   = flag.String("uri", "unix:///var/run/docker.sock", "Docker URI")
	buildFile   = flag.String("f", defaultBuildConfigName, "Build config file")
	buildTarget = flag.String("t", "", "Build target [build|artifact|publish]")
	notify      = flag.Bool("n", false, `Enable notifications (default "false")`)
	showVersion = flag.Bool("version", false, "Show version")

	//buildCfg *BuildConfig
	target byte
)

// VERSION number
const VERSION = "0.1.0"

var (
	branch    string
	commit    string
	buildtime string
)

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if *showVersion {
		printVersion()
		os.Exit(0)
	}
}

func main() {
	buildCfg, err := readBuildConfig(*buildFile)
	if err != nil {
		log.Fatal(err)
	}

	bldr, err := NewDockerWorker(nil)
	if err != nil {
		log.Fatal(err)
	}

	lc := NewLifeCycle(bldr)

	switch *buildTarget {
	case "build":
		err = lc.RunTarget(buildCfg, lifeCycleBuild)
	case "artifact":
		err = lc.RunTarget(buildCfg, lifeCyleArtifacts)
	case "publish":
		err = lc.RunTarget(buildCfg, lifeCyclePublish)
	case "":
		err = lc.Run(buildCfg)
	default:
		err = fmt.Errorf("Invalid target: %s", *buildTarget)
	}

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("")
	}
}
