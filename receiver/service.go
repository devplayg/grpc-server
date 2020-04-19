package receiver

import (
	"context"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
)

type grpcService struct {
	classifier *classifier
	Log        *logrus.Logger
}

func (s *grpcService) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	//p, _ := peer.FromContext(ctx)
	//r.Log.WithFields(logrus.Fields{
	//	"riskLevel": req.Header.RiskLevel,
	//	"client": p.Addr.String(),
	//}).Debug("called")
	_, err := s.classifier.clientApi.Send(context.Background(), req)
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
