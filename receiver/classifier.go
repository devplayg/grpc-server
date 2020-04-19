package receiver

import (
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type classifier struct {
	gRpcClient proto.EventServiceClient
	address    string
	log        *logrus.Logger
	conn       *grpc.ClientConn
	clientApi  proto.EventServiceClient
}

func newClassifier(addr string, log *logrus.Logger) *classifier {
	return &classifier{
		address: addr,
		log:     log,
	}
}

func (c *classifier) connect() error {
	conn, err := grpc.Dial(
		c.address,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&grpc_server.ConnStatsHandler{
			To:  "classifier",
			Log: c.log,
		}),
	)
	if err != nil {
		return err
	}
	c.conn = conn

	// Create client API for service
	c.clientApi = proto.NewEventServiceClient(c.conn)
	return nil
}

func (c *classifier) disconnect() error {
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}
