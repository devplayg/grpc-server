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
	notifierClient *notifierClient
	classifier     *Classifier
	once           sync.Once
	//ch             chan bool
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	s.once.Do(func() {
		// Initial processing time
		grpc_server.ServerStats.Get(grpc_server.StatsInitialProcessing).(*expvar.Int).Set(time.Now().UnixNano())
	})

	grpc_server.ServerStats.Add(grpc_server.StatsWorker, 1)
	go func() {
		defer func() {
			grpc_server.ServerStats.Get(grpc_server.StatsLastProcessing).(*expvar.Int).Set(time.Now().UnixNano())
			grpc_server.ServerStats.Add(grpc_server.StatsWorker, -1)
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
	stats := grpc_server.GetServerStats()
	stats.Meta = fmt.Sprintf("%d", grpc_server.ServerStats.Get(statsInsertingTime).(*expvar.Int).Value())
	return stats, nil
}
