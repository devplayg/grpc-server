package classifier

import (
	"context"
	"expvar"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"sync"
	"time"
)

type grpcService struct {
	notifier   *notifier
	classifier *Classifier
	once       sync.Once
	ch         chan bool
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	s.once.Do(func() {
		stats.Get("start").(*expvar.Int).Set(time.Now().UnixNano())
	})

	s.ch <- true
	stats.Add("worker", 1)
	go func() {
		defer func() {
			stats.Get("end").(*expvar.Int).Set(time.Now().UnixNano())
			stats.Add("worker", -1)
			<-s.ch
		}()

		if err := s.classifier.save(req); err != nil {
			log.Error(err)
			return
		}
	}()
	return &proto.Response{}, nil
}

func (s *grpcService) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}

func (s *grpcService) ResetStats(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	resetStats()
	return &empty.Empty{}, nil
}
