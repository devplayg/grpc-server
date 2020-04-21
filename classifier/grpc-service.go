package classifier

import (
	"context"
	"github.com/devplayg/grpc-server/proto"
	"sync"
	"time"
)

type grpcService struct {
	notifier *notifier
	eventCh  chan<- *proto.Event
	once     sync.Once
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	s.once.Do(func() {
		stats.Set("start", time.Now())
	})

	s.eventCh <- req
	//p, _ := peer.FromContext(ctx)
	//s.log.WithFields(logrus.Fields{
	//	"eventType": req.Header.EventType,
	//	//	"client": p.Addr.String(),
	//}).Debug("received")
	//_, err := s.notifier.clientApi.Send(context.Background(), req)
	//if err != nil {
	//	// Save into file
	//}

	//return nil, status.Errorf(codes.OutOfRange, "err")
	// status.Error(codes.NotFound, "id was not found")
	//return nil, err

	stats.Set("end", time.Now())
	return &proto.Response{}, nil
}

func (s *grpcService) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}
