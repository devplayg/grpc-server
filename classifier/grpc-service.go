package classifier

import (
	"context"
	"expvar"
	"github.com/devplayg/grpc-server/proto"
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
	go func() {
		defer func() {
			stats.Get("end").(*expvar.Int).Set(time.Now().UnixNano())
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
