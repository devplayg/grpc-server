package receiver

import (
	"crypto/tls"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type classifier struct {
	gRpcClient proto.EventServiceClient
	address    string
	log        *logrus.Logger
	conn       *grpc.ClientConn
	clientApi  proto.EventServiceClient
}

func newClassifier(addr string) *classifier {
	return &classifier{
		address: addr,
	}
}

func (c *classifier) connect() error {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			// Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			// Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			// PermitWithoutStream: true,             // send pings even without active streams
		}),
		grpc.WithStatsHandler(&grpc_server.ConnStatsHandler{
			To:  "classifier",
			Log: log,
		}),
	}

	conn, err := grpc.Dial(c.address, opts...)
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
