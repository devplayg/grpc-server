package receiver

import (
	"context"
	"expvar"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

type grpcService struct {
	classifierClient *classifierClient
	storageCh        chan<- *proto.Event
	once             sync.Once
	ch               chan bool
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	s.once.Do(func() {
		// Initial processing time
		grpc_server.ServerStats.Get(grpc_server.StatsInitialProcessing).(*expvar.Int).Set(time.Now().UnixNano())
	})

	s.ch <- true
	grpc_server.ServerStats.Add(grpc_server.StatsWorker, 1)
	go func() {
		defer func() {
			// Last  processing time
			grpc_server.ServerStats.Get(grpc_server.StatsLastProcessing).(*expvar.Int).Set(time.Now().UnixNano())
			grpc_server.ServerStats.Add(grpc_server.StatsWorker, -1)
			<-s.ch
		}()
		if err := s.relayToClassifier(req); err != nil {
			s.storageCh <- req
			log.Error(fmt.Errorf("failed to relay request to classifier; %w", err))
			return
		}

		var size int64
		for _, f := range req.Body.Files {
			size += int64(len(f.Data))
		}
		grpc_server.ServerStats.Add(grpc_server.StatsSize, size)
		grpc_server.ServerStats.Add(grpc_server.StatsCount, 1)
	}()

	return &proto.Response{}, nil
}

func (s *grpcService) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}

func (s *grpcService) ResetDebug(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	grpc_server.ResetServerStats()
	return s.classifierClient.api.ResetDebug(context.Background(), &empty.Empty{})
}

func (s *grpcService) GetServerStats(ctx context.Context, req *empty.Empty) (*proto.ServerStats, error) {
	classifierStats, err := s.classifierClient.api.GetServerStats(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "method Send not implemented")
	}
	stats := grpc_server.GetServerStats()
	stats.Meta = fmt.Sprintf("%d\t%d\t%d\t%d\t%s",
		classifierStats.Count,
		(classifierStats.EndTimeUnixNano-classifierStats.StartTimeUnixNano)/int64(time.Millisecond),
		classifierStats.Size,
		classifierStats.TotalWorkingTimeMilli,
		classifierStats.Meta,
	)

	return stats, nil
}

func (s *grpcService) relayToClassifier(req *proto.Event) error {
	if s.classifierClient.conn.GetState() != connectivity.Ready {
		return fmt.Errorf("connection is not ready")
	}
	_, err := s.classifierClient.api.Send(context.Background(), req)
	return err
}
