package receiver

import (
	"context"
	"expvar"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
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

	stats.Add("worker", 1)
	go func() {
		log.Trace("sending")
		defer func() {
			log.Trace("done")
			stats.Add("worker", -1)
		}()
		if err := s.relayToClassifier(req); err != nil {
			s.storageCh <- req
			log.Error("failed to relay request  to classifier")
			return
		}

		var size int64
		for _, f := range req.Body.Files {
			size += int64(len(f.Data))
		}

		stats.Add("size", size)
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

func (s *grpcService) ResetDebug(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	resetStats()
	return s.classifier.clientApi.ResetDebug(context.Background(), &empty.Empty{})
}

func (s *grpcService) Debug(ctx context.Context, req *empty.Empty) (*proto.DebugMessage, error) {
	debug, err := s.classifier.clientApi.Debug(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "method Send not implemented")
	}

	duration := (stats.Get("end").(*expvar.Int).Value() - stats.Get("start").(*expvar.Int).Value()) / int64(time.Millisecond)
	relayed := stats.Get("relayed").(*expvar.Int).Value()
	size := stats.Get("size").(*expvar.Int).Value()

	str := fmt.Sprintf("%d\t%d\t%d", relayed, duration, size)
	return &proto.DebugMessage{
		Message: fmt.Sprintf("%s\t%s", str, debug.Message),
	}, nil
}

func (s *grpcService) relayToClassifier(req *proto.Event) error {
	if s.classifier.conn.GetState() != connectivity.Ready {
		return fmt.Errorf("connection is not ready")
	}
	_, err := s.classifier.clientApi.Send(context.Background(), req)
	return err
}
