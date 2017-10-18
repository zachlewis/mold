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
	dockerURI   = flag.String("uri", "", "Docker URI")
	buildFile   = flag.String("f", defaultBuildConfigName, "Build config file")
	buildTarget = flag.String("t", "", "Build target [build|artifacts|publish]")

	showVersion = flag.Bool("version", false, "Show version")
	variable    = flag.String("var", "", "Show value of vairable specified in the configuration file")

	initMoldCfg    = flag.Bool("init", false, "Initialize a new mold file.")
	showAppVersion = flag.Bool("app-version", false, "Show the app version per mold")
)

func init() {
	flag.Usage = printUsage
	flag.Parse()
	//log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func initializeBuild(moldFile, uri string) (*MoldConfig, *DockerWorker, error) {
	moldConfig, err := readMoldConfig(moldFile)
	if err == nil {
		var dcli *Docker
		if dcli, err = NewDocker(uri); err == nil {
			var dw *DockerWorker
			if dw, err = NewDockerWorker(dcli); err == nil {
				return moldConfig, dw, nil
			}
		}
	}
	return nil, nil, err
}

func getVar(key, moldFile string) (string, error) {
	moldConfig, err := readMoldConfig(moldFile)
	if err == nil {
		if len(moldConfig.Variables) > 0 {
			if val, ok := moldConfig.Variables[key]; ok {
				return val, nil
			}
		}
		return "", fmt.Errorf("Variable not specified")
	}
	return "", err
}

func main() {
	if *showVersion {
		printVersion()
		os.Exit(0)
	} else if *showAppVersion {
		gt, _ := newGitVersion(".")
		fmt.Println(gt.Version())
		os.Exit(0)
	} else if *initMoldCfg {
		if err := initializeMoldConfig("."); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(*variable) > 0 {
		val, err := getVar(*variable, *buildFile)
		if err == nil {
			fmt.Printf("%s\n", val)
		} else {
			log.Println("ERR", err)
		}
		os.Exit(0)
	}

	moldConfig, worker, err := initializeBuild(*buildFile, *dockerURI)
	if err != nil {
		log.Fatal(err)
	}

	lc := NewLifeCycle(worker)
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
		err = lc.Run(moldConfig)
	default:
		if targetArg == "" {
			err = lc.RunTarget(moldConfig, target)
		} else {
			err = lc.RunTarget(moldConfig, target, targetArg)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
}
