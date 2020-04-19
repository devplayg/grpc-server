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
	config         *grpc_server.Config
	gRpcServer     *grpc.Server
	gRpcClientConn *grpc.ClientConn
}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}

	// Connect to classifier
	gRpcClient, err := r.connectToClassifier()
	if err != nil {
		return fmt.Errorf("failed to connect to classifier; %w", err)
	}

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := r.startGrpcServer(gRpcClient); err != nil {
			r.Log.Errorf("failed to start gRPC server: %w", err)
			return
		}
		r.Log.Info("gRpcServer has been stopped")
	}()

	<-r.Ctx.Done()

	// Stop gRPC server
	r.gRpcServer.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	defer r.Log.Infof("%s has been stopped", r.Engine.Config.Name)

	if r.gRpcClientConn != nil {
		//r.Log.Debugf("connection status=%v", r.gRpcClientConn.GetState())
		if err := r.gRpcClientConn.Close(); err != nil {
			return err
		}
	}

	return nil
}
