package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/devplayg/hippo/v2"
	"google.golang.org/grpc"
	"net"
)

// Classifier receives data from receiver via gRPC framework
type Classifier struct {
	hippo.Launcher
	config         *grpc_server.Config
	gRpcServer     *grpc.Server
	gRpcClientConn *grpc.ClientConn
}

func NewClassifier() *Classifier {
	return &Classifier{}
}

func (c *Classifier) Start() error {
	if err := c.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", c.Engine.Config.Name, err)
	}

	// Connect to classifier
	gRpcClient, err := c.connectToNotifier()
	if err != nil {
		return fmt.Errorf("failed to start gRPC sender")
	}

	ch := make(chan bool)
	go func() {
		if err := c.startGRPCServer(); err != nil {
			c.Log.Errorf("failed to start gRPC server: %w", err)
		}
		c.Log.Debug("gRpcServer has been stopped")
		close(ch)
	}()

	<-c.Ctx.Done()

	// Stop gRPC server
	c.gRpcServer.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (c *Classifier) Stop() error {
	c.Log.Infof("%s has been stopped", c.Engine.Config.Name)
	return nil
}

func (c *Classifier) startGRPCServer() error {
	ln, err := net.Listen("tcp", c.config.App.Classifier.Address)
	if err != nil {
		return err
	}

	// Create gRPC server
	c.gRpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.UnaryInterceptor),
		grpc.StatsHandler(&grpc_server.ConnStatsHandler{
			From: "receiver",
			To:   "classifier",
			Log:  c.Log,
		}),
	)

	// Register server to gRPC server
	proto.RegisterEventServiceServer(c.gRpcServer, &eventReceiver{})

	// Run
	c.Log.Debug("gRpcServer has been started")
	if err := c.gRpcServer.Serve(ln); err != nil {
		return err
	}
	return nil
}
