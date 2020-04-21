package receiver

import (
	"context"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/connectivity"
	"time"
)

type grpcService struct {
	classifier *classifier
	storageCh  chan<- *proto.Event
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	// p, _ := peer.FromContext(ctx)
	log.WithFields(logrus.Fields{
		"eventType": req.Header.EventType,
		// "client": p.Addr.String(),
	}).Trace("received")

	go func() {
		started := time.Now()
		if err := s.relayToClassifier(req); err != nil {
			s.storageCh <- req
			log.Error("failed to relay request  to classifier")
			return
		}
		stats.Add("relayed", 1)
		stats.Add("relayed-time", time.Since(started).Milliseconds())
	}()

	// return nil, status.Errorf(codes.OutOfRange, "err")
	// status.Error(codes.NotFound, "id was not found")
	// return nil, err

	return &proto.Response{}, nil
}

func (s *grpcService) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}

func (s *grpcService) relayToClassifier(req *proto.Event) error {
	if s.classifier.conn.GetState() != connectivity.Ready {
		return fmt.Errorf("connection is not ready")
	}
	_, err := s.classifier.clientApi.Send(context.Background(), req)
	return err
}
