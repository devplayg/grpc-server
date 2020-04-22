package classifier

import (
	"context"
	"github.com/devplayg/grpc-server/proto"
)

type grpcService struct {
	notifier *notifier
	//eventCh  chan<- *proto.Event
	classifier *Classifier
	//once     sync.Once
	ch chan bool
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	s.ch <- true
	go func() {
		defer func() {
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
