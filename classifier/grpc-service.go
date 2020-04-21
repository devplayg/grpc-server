package classifier

import (
	"context"
	"github.com/devplayg/grpc-server/proto"
	"sync"
)

type grpcService struct {
	notifier *notifier
	eventCh  chan<- *proto.Event
	once     sync.Once
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	s.eventCh <- req
	return &proto.Response{}, nil
}

func (s *grpcService) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}
