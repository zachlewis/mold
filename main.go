package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	dockerURI   = flag.String("uri", "unix:///var/run/docker.sock", "Docker URI")
	buildFile   = flag.String("f", defaultBuildConfigName, "Build config file")
	buildTarget = flag.String("t", "", "Build target [build|artifacts|publish]")
	//notify      = flag.Bool("n", false, `Enable notifications (default "false")`)
	showVersion = flag.Bool("version", false, "Show version")
)

func init() {
	flag.Usage = printUsage
	flag.Parse()
	//log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func initializeBuild(bldfile, uri string) (*BuildConfig, *DockerWorker, error) {
	buildCfg, err := readBuildConfig(bldfile)
	if err == nil {
		var dcli *Docker
		if dcli, err = NewDocker(uri); err == nil {
			var bldr *DockerWorker
			if bldr, err = NewDockerWorker(dcli); err == nil {
				return buildCfg, bldr, nil
			}
		}
	}
	return nil, nil, err
}

func main() {
	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	buildCfg, bldr, err := initializeBuild(*buildFile, *dockerURI)
	if err != nil {
		log.Fatal(err)
	}

	lc := NewLifeCycle(bldr)
	// Listen for signals for a clean shutdown
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		if e := lc.Abort(); e != nil {
			log.Println("ERR", e)
		}
	}()
	// Run targets
	target, targetArg := parseTarget(*buildTarget)
	switch target {
	case "":
		err = lc.Run(buildCfg)
	default:
		err = lc.RunTarget(buildCfg, target, targetArg)
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
}
