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
	fs       = pflag.NewFlagSet(appName, pflag.ContinueOnError)
	debug    = fs.Bool("debug", false, "Debug")
	verbose  = fs.BoolP("verbose", "v", false, "Verbose")
	version  = fs.Bool("version", false, "Version")
	insecure = fs.Bool("insecure", false, "Disable TLS")
	certFile = fs.String("certFile", "server.crt", "SSL Certificate file")
	keyFile  = fs.String("keyFile", "server.key", "SSL Certificate key file")
)

func main() {
	config := hippo.Config{
		Name:        appName,
		Description: appDescription,
		Version:     appVersion,
		Debug:       *debug,
		CertFile:    *certFile,
		KeyFile:     *keyFile,
		Insecure:    *insecure,
		LogDir:      ".",
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
