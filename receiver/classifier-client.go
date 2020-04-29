package receiver

import (
	"crypto/tls"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"time"
)

type classifierClient struct {
	gRpcClient proto.EventServiceClient
	address    string
	log        *logrus.Logger
	conn       *grpc.ClientConn
	api        proto.EventServiceClient
	insecure   bool
}

func newClassifierClient(addr string, insecure bool) *classifierClient {
	return &classifierClient{
		address:  addr,
		insecure: insecure,
	}
}

func (c *classifierClient) connect() error {
	conn, err := grpc.Dial(c.address, c.getGrpcDialOptions()...)
	if err != nil {
		return err
	}

	// Create connection
	c.conn = conn

	// Create client API for service
	c.api = proto.NewEventServiceClient(c.conn)
	return nil
}

func (c *classifierClient) getGrpcDialOptions() []grpc.DialOption {
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

	// Option for insecure mode
	if !c.insecure {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Backoff
	connParam := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  1.0 * time.Second,
			Multiplier: 1.6,
			Jitter:     0.2,
			MaxDelay:   5.0 * time.Second,
		},
		MinConnectTimeout: 1 * time.Second,
	}
	opts = append(opts, grpc.WithConnectParams(connParam))

	return opts
}

func (c *classifierClient) disconnect() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
