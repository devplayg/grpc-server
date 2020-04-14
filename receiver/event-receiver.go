package receiver

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/devplayg/grpc-server/proto"
)

type eventReceiver struct {
}

func (r *eventReceiver) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	spew.Dump(req)
	return &proto.Response{
		Error: "",
	}, nil
}

func (r *eventReceiver) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{
		Error: "",
	}, nil
}
