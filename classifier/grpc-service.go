package classifier

import (
	"context"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
)

type grpcService struct {
	notifier *notifier
	log      *logrus.Logger
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	//p, _ := peer.FromContext(ctx)
	s.log.WithFields(logrus.Fields{
		"riskLevel": req.Header.RiskLevel,
		//	"client": p.Addr.String(),
	}).Debug("received")
	_, err := s.notifier.clientApi.Send(context.Background(), req)
	if err != nil {
		// Save into file
	}

	//return nil, status.Errorf(codes.OutOfRange, "err")
	// status.Error(codes.NotFound, "id was not found")
	//return nil, err

	return &proto.Response{}, nil
}

func (s *grpcService) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}
