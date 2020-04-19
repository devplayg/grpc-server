package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"math/rand"
	"os"

	"time"
)

const addr = "localhost:8801"

func init() {
	rand.Seed(time.Now().UnixNano())
	_ = fs.Parse(os.Args[1:])
}

func main() {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := []grpc.DialOption{
		// grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			// Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			// Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			// PermitWithoutStream: true,             // send pings even without active streams
		}),
	}

	// Create connection
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		panic(err)
	}

	// Create client API for service
	clientApi := proto.NewEventServiceClient(conn)

	// gRPC remote procedure call

	for {
		event := generateEvent()
		_, err = clientApi.Send(context.Background(), event)
		if err != nil {
			statusErr, ok := status.FromError(err)
			if !ok {
				panic(err)
			}
			fmt.Printf("[error] %v\n", statusErr.Message())
			fmt.Printf("[error] %v\n", statusErr.Code())
			fmt.Printf("[error] %v\n", statusErr.Details())
			fmt.Printf("[error] %v\n", statusErr.Err())
		}

		time.Sleep(5 * time.Second)
	}

	//time.Sleep(3000 * time.Millisecond)
	//}
	//select {}
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
