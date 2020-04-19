package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/hippo/v2"
	"google.golang.org/grpc"
)

// Classifier receives data from receiver via gRPC framework
type Classifier struct {
	hippo.Launcher
	config         *grpc_server.Config
	gRpcServer     *grpc.Server
	gRpcClientConn *grpc.ClientConn
	notifier       *notifier
}

func NewClassifier() *Classifier {
	return &Classifier{}
}

func (c *Classifier) Start() error {
	if err := c.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", c.Engine.Config.Name, err)
	}

	// Connect to classifier
	c.notifier = newNotifier(c.config.App.Classifier.Notifier.Address, c.Log)
	if err := c.notifier.connect(); err != nil {
		return fmt.Errorf("failed to connect to notifier: %w", err)
	}

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := c.startGrpcServer(); err != nil {
			c.Log.Error(fmt.Errorf("failed to start gRPC server: %w", err))
			return
		}
		c.Log.Debug("gRpcServer has been stopped")
	}()
	c.Log.Infof("%s has been started", c.Engine.Config.Name)

	<-c.Ctx.Done()

	// Stop gRPC server
	c.gRpcServer.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (c *Classifier) Stop() error {
	defer c.Log.Infof("%s has been stopped", c.Engine.Config.Name)

	if err := c.notifier.disconnect(); err != nil {
		c.Log.Error("failed to disconnect classifier")
	}
	return nil
}
