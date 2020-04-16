package main

import (
	"context"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"math/rand"
	"time"
)

const addr = "localhost:8801"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Create connection
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create client API for service
	clientApi := proto.NewEventServiceClient(conn)

	// gRPC remote procedure call
	//for {
	event := generateEvent()
	// spew.Dump(event)
	res, err := clientApi.Send(context.Background(), event)
	if err != nil {
		fmt.Printf("[error] %s\n", err.Error())
	}
	if res != nil {
		if len(res.Error) > 0 {
			fmt.Printf("[error] %s\n", res.Error)
		}
	}
	time.Sleep(5000 * time.Millisecond)
	//}
}

func generateEvent() *proto.Event {
	now := time.Now()
	t := &timestamp.Timestamp{
		Seconds: now.Unix(),
		Nanos:   int32(now.Nanosecond()),
	}

	return &proto.Event{
		Header: &proto.EventHeader{
			Version:   1,
			Date:      t,
			RiskLevel: proto.EventHeader_RiskLevel(rand.Intn(5) + 1),
		},
		Body: &proto.EventBody{
			Files: []*proto.File{
				{
					Time:     t,
					Category: rand.Int31n(5) + 1,
					Data:     []byte("abc"),
				},
			},
		},
	}
}
