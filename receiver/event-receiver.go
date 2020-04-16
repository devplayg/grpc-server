package receiver

import (
	"context"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"google.golang.org/grpc/peer"
)

type eventReceiver struct {
	grpcSender proto.EventServiceClient
}

func (r *eventReceiver) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	p, ok := peer.FromContext(ctx)
	if ok {
		fmt.Printf("requested by %s\n", p.Addr.String())
	}
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
	return &proto.Response{
		Error: "",
	}, nil
}

func (r *eventReceiver) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{
		Error: "",
	}, nil
}
