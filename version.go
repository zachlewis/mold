package main

import "fmt"

// VERSION number
const VERSION = "0.1.0"

func printVersion() {
	fmt.Printf("%s commit=%s branch=%s buildtime=%s\n", VERSION, commit, branch, buildtime)
}
