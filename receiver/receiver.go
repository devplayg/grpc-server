package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/devplayg/hippo/v2"
	"google.golang.org/grpc"
	"net"
)

type Receiver struct {
	hippo.Launcher
	config     *grpc_server.Config
	gRpcServer *grpc.Server
}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}
	r.Log.Infof("%s has been started", r.Engine.Config.Name)

	ch := make(chan bool)
	go func() {
		if err := r.startGRPCServer(); err != nil {
			r.Log.Errorf("failed to start gRPC server: %w", err)
		}
		r.Log.Debug("gRpcServer has been stopped")
		close(ch)
	}()

	<-r.Ctx.Done()

	// Stop gRPC server
	r.gRpcServer.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	r.Log.Infof("%s has been stopped", r.Engine.Config.Name)
	return nil
}

func (r *Receiver) startGRPCServer() error {
	ln, err := net.Listen("tcp", r.config.App.Receiver.BindAddress)
	if err != nil {
		return err
	}

	// Create gRPC server
	r.gRpcServer = grpc.NewServer()

	// Register server to gRPC server
	proto.RegisterEventServiceServer(r.gRpcServer, &eventReceiver{})

	// Run
	r.Log.Debug("gRpcServer has been started")
	if err := r.gRpcServer.Serve(ln); err != nil {
		return err
	}
	return nil
}
