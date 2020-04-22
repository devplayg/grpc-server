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
	insecure   bool
}

func newClassifier(addr string, insecure bool) *classifier {
	return &classifier{
		address:  addr,
		insecure: insecure,
	}
}

func (c *classifier) connect() error {
	opts := []grpc.DialOption{
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

	if !c.insecure {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		opts = append(opts, grpc.WithInsecure())
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
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
