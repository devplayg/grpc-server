package receiver

import (
	"context"
	"expvar"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"google.golang.org/grpc/connectivity"
	"sync"
	"time"
)

type grpcService struct {
	classifier *classifier
	storageCh  chan<- *proto.Event
	once       sync.Once
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	s.once.Do(func() {
		stats.Get("start").(*expvar.Int).Set(time.Now().UnixNano())
	})

	go func() {
		if err := s.relayToClassifier(req); err != nil {
			s.storageCh <- req
			log.Error("failed to relay request  to classifier")
			return
		}
		for _, f := range req.Body.Files {
			stats.Add("size", int64(len(f.Data)))
		}
		stats.Add("relayed", 1)
		stats.Get("end").(*expvar.Int).Set(time.Now().UnixNano())
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
