package classifier

import (
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"google.golang.org/grpc"
)

func (c *Classifier) connectToNotifier() (proto.EventServiceClient, error) {
	notifier, err := grpc.Dial(
		c.config.App.Classifier.Notifier.Address,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&grpc_server.ConnStatsHandler{
			To:  "classifier",
			Log: c.Log,
		}),
	)
	if err != nil {
		return nil, err
	}
	c.gRpcClientConn = notifier

	// Create client API for service
	classifierApi := proto.NewEventServiceClient(c.gRpcClientConn)
	return classifierApi, nil
}
