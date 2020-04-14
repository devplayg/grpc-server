package main

import (
	"fmt"
	"github.com/devplayg/grpc-server/classifier"
	"github.com/devplayg/hippo/v2"
	"github.com/spf13/pflag"
	"os"
)

const (
	appName        = "classifier"
	appDescription = "Classifier via gRPC"
	appVersion     = "1.0.0"
)

var (
	fs      = pflag.NewFlagSet(appName, pflag.ContinueOnError)
	debug   = fs.Bool("debug", false, "Debug")
	verbose = fs.BoolP("verbose", "v", false, "Verbose")
	version = fs.Bool("version", false, "Version")
)

func main() {
	config := hippo.Config{
		Name:        appName,
		Description: appDescription,
		Version:     appVersion,
		Debug:       *debug,
		// LogDir:      ".",
	}
	if *verbose {
		config.LogDir = ""
	}

	engine := hippo.NewEngine(classifier.NewClassifier(), &config)
	if err := engine.Start(); err != nil {
		panic(err)
	}
}

func init() {
	_ = fs.Parse(os.Args[1:])

	if *version {
		fmt.Printf("%s %s\n", appName, appVersion)
		os.Exit(0)
	}
}
