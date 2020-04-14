package receiver

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/devplayg/grpc-server/proto"
)

type eventReceiver struct {
	grpcSender proto.EventServiceClient
}

func (r *eventReceiver) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	spew.Dump(req)
	res, err := r.grpcSender.Send(context.Background(), req)
	if err != nil {
		fmt.Printf("[error] %w\n", err.Error())
	}
	spew.Dump(res)

	return &proto.Response{
		Error: "",
	}, nil
}

func (r *eventReceiver) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{
		Error: "",
	}, nil
}
