package main

import (
	"fmt"
	"github.com/devplayg/grpc-server/receiver"
	"github.com/devplayg/hippo/v2"
	"github.com/spf13/pflag"
	"os"
	"time"
)

const (
	appName        = "receiver"
	appDescription = "Receiver via gRPC"
	appVersion     = "1.0.0"
)

var (
	fs           = pflag.NewFlagSet(appName, pflag.ExitOnError)
	debug        = fs.Bool("debug", false, "Debug") // GODEBUG=http2debug=2
	trace        = fs.Bool("trace", false, "Trace")
	verbose      = fs.BoolP("verbose", "v", false, "Verbose")
	version      = fs.Bool("version", false, "Version")
	insecure     = fs.Bool("insecure", false, "Disable TLS")
	certFile     = fs.String("certFile", "server.crt", "SSL Certificate file")
	keyFile      = fs.String("keyFile", "server.key", "SSL Certificate key file")
	batchSize    = fs.Int("batchsize", 10000, "Batch size")
	batchTimeout = fs.Int("batchtime", 1000, "Batch timeout, in milliseconds")
	storage      = fs.String("storage", "data", "Storage path")
)

func main() {
	config := hippo.Config{
		Name:        appName,
		Description: appDescription,
		Version:     appVersion,
		Debug:       *debug,
		Trace:       *trace,
		CertFile:    *certFile,
		KeyFile:     *keyFile,
		Insecure:    *insecure,
		LogDir:      ".",
	}
	if *verbose {
		config.LogDir = ""
	}

	server := receiver.NewReceiver(*batchSize, time.Duration(*batchTimeout)*time.Millisecond, *storage)
	engine := hippo.NewEngine(server, &config)
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
