package main

import (
	"fmt"
	"github.com/devplayg/grpc-server/receiver"
	"github.com/devplayg/hippo/v2"
	"github.com/spf13/pflag"
	"os"
)

const (
	appName        = "receiver"
	appDescription = "Receiver via gRPC"
	appVersion     = "1.0.0"
)

var (
	fs      = pflag.NewFlagSet(appName, pflag.ContinueOnError)
	debug   = fs.Bool("debug", true, "Debug")
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

	engine := hippo.NewEngine(receiver.NewReceiver(), &config)
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

//
//
//func Path(rel string) string {
//	if filepath.IsAbs(rel) {
//		return rel
//	}
//
//	return filepath.Join(basepath, rel)
//}
