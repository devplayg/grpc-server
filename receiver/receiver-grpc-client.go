package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"google.golang.org/grpc"
)

func (r *Receiver) connectWithClassifier() (proto.EventServiceClient, error) {
	classifierConn, err := grpc.Dial(
		r.config.App.Receiver.Classifier.Address,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&grpc_server.ConnStatsHandler{
			To:  fmt.Sprintf("%s", "classifier"),
			Log: r.Log,
		}),
	)
	if err != nil {
		return nil, err
	}
	r.classifierConn = classifierConn

	// Create client API for service
	classifierApi := proto.NewEventServiceClient(r.classifierConn)
	return classifierApi, nil
}
