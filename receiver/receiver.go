package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/hippo/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

var log *logrus.Logger

// Receiver receives data from agents via gRPC framework
type Receiver struct {
	hippo.Launcher
	config     *grpc_server.Config
	gRpcServer *grpc.Server
	classifier *classifier
	assistant  *assistant
}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}

	// Classifier
	r.classifier = newClassifier(r.config.App.Receiver.Classifier.Address)
	if err := r.classifier.connect(); err != nil {
		return fmt.Errorf("failed to connect to classifier: %w", err)
	}

	// Start assistant
	r.assistant = newAssistant(10, 5*time.Second)
	r.assistant.Start()

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := r.startGrpcServer(r.assistant.ch); err != nil {
			log.Error(fmt.Errorf("failed to start gRPC server: %w", err))
			return
		}
		log.Info("gRpcServer has been stopped")
	}()
	log.Infof("%s has been started", r.Engine.Config.Name)

	<-r.Ctx.Done()

	// Stop gRPC server
	r.gRpcServer.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	defer log.Infof("%s has been stopped", r.Engine.Config.Name)

	if err := r.classifier.disconnect(); err != nil {
		log.Error("failed to disconnect classifier")
	}

	return nil
}
