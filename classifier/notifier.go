package classifier

import (
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"google.golang.org/grpc"
)

type notifier struct {
	gRpcClient proto.EventServiceClient
	address    string
	conn       *grpc.ClientConn
	clientApi  proto.EventServiceClient
}

func newNotifier(addr string) *notifier {
	return &notifier{
		address: addr,
	}
}

func (n *notifier) connect() error {
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

func (n *notifier) disconnect() error {
	if err := n.conn.Close(); err != nil {
		return err
	}
	return nil
}
