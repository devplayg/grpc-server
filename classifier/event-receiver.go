package classifier

import (
	"context"
	"github.com/devplayg/grpc-server/proto"
)

type eventReceiver struct {
}

func (r *eventReceiver) Send(ctx context.Context, req *proto.Event) (*proto.Response, error) {
	return &proto.Response{}, nil
}

func (r *eventReceiver) SendHeader(ctx context.Context, req *proto.EventHeader) (*proto.Response, error) {
	return &proto.Response{}, nil
}
