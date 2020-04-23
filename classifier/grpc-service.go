package classifier

import (
	"context"
	"expvar"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"sync"
	"time"
)

const statsInsertingTime = "inserting-time"

type grpcService struct {
	notifier   *notifier
	classifier *Classifier
	once       sync.Once
	ch         chan bool
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
			grpc_server.ServerStats.Get(grpc_server.StatsLastProcessing).(*expvar.Int).Set(time.Now().UnixNano())
			grpc_server.ServerStats.Add(grpc_server.StatsWorker, -1)
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

func (s *grpcService) ResetDebug(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	grpc_server.ResetServerStats(statsInsertingTime)
	return &empty.Empty{}, nil
}

func (s *grpcService) GetServerStats(ctx context.Context, req *empty.Empty) (*proto.ServerStats, error) {
	return &proto.ServerStats{
		StartTimeUnixNano: grpc_server.ServerStats.Get(grpc_server.StatsInitialProcessing).(*expvar.Int).Value(),
		EndTimeUnixNano:   grpc_server.ServerStats.Get(grpc_server.StatsLastProcessing).(*expvar.Int).Value(),
		Count:             grpc_server.ServerStats.Get(grpc_server.StatsCount).(*expvar.Int).Value(),
		Size:              grpc_server.ServerStats.Get(grpc_server.StatsSize).(*expvar.Int).Value(),
		Worker:            int32(grpc_server.ServerStats.Get(grpc_server.StatsWorker).(*expvar.Int).Value()),
		Meta:              fmt.Sprintf("%d", grpc_server.ServerStats.Get(statsInsertingTime).(*expvar.Int).Value()),
	}, nil
}
