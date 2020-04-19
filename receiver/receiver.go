package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/hippo/v2"
	"google.golang.org/grpc"
)

// Receiver receives data from agents via gRPC framework
type Receiver struct {
	hippo.Launcher
	config     *grpc_server.Config
	gRpcServer *grpc.Server
	classifier *classifier
}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}

	r.classifier = newClassifier(r.config.App.Receiver.Classifier.Address, r.Log)
	if err := r.classifier.connect(); err != nil {
		return fmt.Errorf("failed to connect to classifier: %w", err)
	}

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := r.startGrpcServer(); err != nil {
			r.Log.Error(fmt.Errorf("failed to start gRPC server: %w", err))
			return
		}
		r.Log.Info("gRpcServer has been stopped")
	}()
	r.Log.Infof("%s has been started", r.Engine.Config.Name)

	<-r.Ctx.Done()

	// Stop gRPC server
	r.gRpcServer.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	defer r.Log.Infof("%s has been stopped", r.Engine.Config.Name)

	if err := r.classifier.disconnect(); err != nil {
		r.Log.Error("failed to disconnect classifier")
	}

	return nil
}
