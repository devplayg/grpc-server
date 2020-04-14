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
	config       *grpc_server.Config
	grpcReceiver *grpc.Server
	//grpcSender   proto.EventServiceClient
}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}
	r.Log.Infof("%s has been started", r.Engine.Config.Name)

	grpcSender, err := r.startGRPCSender()
	if err != nil {
		return fmt.Errorf("failed to start gRPC sender")
	}

	ch := make(chan bool)
	go func() {
		if err := r.startGRPCReceiver(grpcSender); err != nil {
			r.Log.Errorf("failed to start gRPC server: %w", err)
		}
		r.Log.Debug("gRpcServer has been stopped")
		close(ch)
	}()

	<-r.Ctx.Done()

	// Stop gRPC server
	r.grpcReceiver.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	r.Log.Infof("%s has been stopped", r.Engine.Config.Name)
	return nil
}

func (r *Receiver) startGRPCReceiver(grpcSender proto.EventServiceClient) error {
	ln, err := net.Listen("tcp", r.config.App.Receiver.Address)
	if err != nil {
		return err
	}

	// Create gRPC server
	r.grpcReceiver = grpc.NewServer()

	// Register server to gRPC server
	proto.RegisterEventServiceServer(r.grpcReceiver, &eventReceiver{grpcSender})

	// Run
	r.Log.Debug("gRpcServer has been started")
	if err := r.grpcReceiver.Serve(ln); err != nil {
		return err
	}
	return nil
}

func (r *Receiver) startGRPCSender() (proto.EventServiceClient, error) {
	conn, err := grpc.Dial(r.config.App.Receiver.Classifier.Address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create client API for service
	grpcSender := proto.NewEventServiceClient(conn)

	// gRPC remote procedure call
	//for {
	//	event := generateEvent()
	//	spew.Dump(event)
	//	_, err := clientApi.Send(context.Background(), event)
	//	if err != nil {
	//		panic(err)
	//	}
	//	time.Sleep(100 * time.Millisecond)
	//
	//}

	return grpcSender, nil
}
