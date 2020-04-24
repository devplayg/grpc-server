package classifier

import (
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"google.golang.org/grpc"
)

type notifierClient struct {
	gRpcClient proto.EventServiceClient
	address    string
	conn       *grpc.ClientConn
	clientApi  proto.EventServiceClient
}

func newNotifier(addr string) *notifierClient {
	return &notifierClient{
		address: addr,
	}
}

func (n *notifierClient) connect() error {
	conn, err := grpc.Dial(
		n.address,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&grpc_server.ConnStatsHandler{
			To:  "classifier",
			Log: log,
		}),
	)
	if err != nil {
		return err
	}
	n.conn = conn

	// Create client API for service
	n.clientApi = proto.NewEventServiceClient(n.conn)
	return nil
}

func (n *notifierClient) disconnect() error {
	if n.conn != nil {
		if err := n.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
