package main

import (
	"flag"
	"fmt"
	"os"
)

const VERSION = "0.1.0"

const usage = "Usage: odu [options] -- <program> [<arg>...]"

var dirFlag = flag.String("dir", ".", "working directory for the spawned process")
var logFlag = flag.String("log", "", "enable logging")
var outputFlag = flag.String("output", "", "path to output fifo")
var versionFlag = flag.Bool("v", false, "print version and exit")

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", VERSION)
		os.Exit(0)
	}

	if pipeExists(*outputFlag) {
		die_usage("output is not a pipe")
	}

	initLogger(*logFlag)

	args := flag.Args()
	validateArgs(args)

	err := executor(*dirFlag, *outputFlag, args)
	if err != nil {
		os.Exit(getExitStatus(err))
	}
}

func validateArgs(args []string) {
	if len(args) < 1 {
		die_usage("Not enough arguments.")
	}

	logger.Printf("Flag values:\n  dir: %v\nArgs: %v\n", *dirFlag, args)
}

func pipeExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.Mode()&os.ModeNamedPipe == 0
}
