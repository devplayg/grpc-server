package receiver

import (
	"context"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
)

type eventReceiver struct {
	grpcSender proto.EventServiceClient
}

func (r *eventReceiver) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	//p, _ := peer.FromContext(ctx)
	//if ok {
	//}
	//fmt.Printf("[requested by %s] risk-level=%d\n", p.Addr.String(), req.Header.RiskLevel)
	//request.Frontendip = p.Addr.String()
	//spew.Dump(req)
	//res, err := r.grpcSender.Send(context.Background(), req)
	//if err != nil {
	//	//fmt.Printf("[error] %w\n", err.Error())
	//	return &proto.Response{
	//		Error: "#1:"+err.Error(),
	//	}, nil
	//}
	//if len(res.Error) < 1 {
	//	return &proto.Response{
	//		Error: "#2:"+err.Error(),
	//	}, nil
	//}
	//fmt.Printf("[requested by %s] risk-level=%d\n", p.Addr.String(), req.Header.RiskLevel)
	return &proto.Response{}, nil
}

func (r *eventReceiver) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}

type output struct {
	grpcSender proto.EventServiceClient
	Log        *logrus.Logger
}

func (r *output) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	//p, _ := peer.FromContext(ctx)
	//r.Log.WithFields(logrus.Fields{
	//	"riskLevel": req.Header.RiskLevel,
	//	"client": p.Addr.String(),
	//}).Debug("called")
	_, err := r.grpcSender.Send(context.Background(), req)
	if err != nil {
		// Save into file
	}

	//return nil, status.Errorf(codes.OutOfRange, "err")
	// status.Error(codes.NotFound, "id was not found")
	//return nil, err

	return &proto.Response{}, nil
}

func (r *output) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}
