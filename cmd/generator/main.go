package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"

	"time"
)

const addr = "localhost:8801"

var (
	fs         = pflag.NewFlagSet("generator", pflag.ContinueOnError)
	agentCount = fs.IntP("agent", "a", 3, "Client count")
	dataCount  = fs.IntP("c", "c", 1, "Event count by client")
	devices    []string
	images     [][]byte
)

func init() {
	rand.Seed(time.Now().UnixNano())
	_ = fs.Parse(os.Args[1:])

	devices = make([]string, 0)
	start := 65
	for i := start; i < start+(*agentCount); i++ {
		devices = append(devices, "DEVICE-"+string(i))
	}

	imgCount := 3
	images = make([][]byte, imgCount)
	for i := 1; i <= imgCount; i++ {
		b, err := ioutil.ReadFile(fmt.Sprintf("sample%d.jpg", i))
		if err != nil {
			panic(err)
		}
		images[i-1] = b
	}
}

func generateData() map[int32][]*proto.Event {
	data := make(map[int32][]*proto.Event)
	for a := 0; a < *agentCount; a++ {
		k := int32(a)
		data[k] = make([]*proto.Event, 0)
		for i := 0; i < *dataCount; i++ {
			data[k] = append(data[k], generateEvent(devices[k]))
		}
	}
	return data
}

func main() {
	data := generateData()
	fmt.Printf("data generated\n")
	wg := new(sync.WaitGroup)
	for i := 0; i < *agentCount; i++ {
		wg.Add(1)
		k := int32(i)
		go send(wg, data[k])
	}
	wg.Wait()
	fmt.Printf("%d sent\n", (*agentCount)*(*dataCount))
}

func send(wg *sync.WaitGroup, events []*proto.Event) {
	defer wg.Done()

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := []grpc.DialOption{
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
	defer conn.Close()

	// Create client API for service
	clientApi := proto.NewEventServiceClient(conn)

	// gRPC remote procedure call
	for _, e := range events {
		_, err = clientApi.Send(context.Background(), e)
		if err != nil {
			fmt.Printf("[error] %s\n", err.Error())
			continue
		}
	}
}

func generateEvent(deviceCode string) *proto.Event {
	now := time.Now()
	t := &timestamp.Timestamp{
		Seconds: now.Unix(),
		Nanos:   int32(now.Nanosecond()),
	}

	//data := make([]byte, 16)
	//rand.Read(data)

	return &proto.Event{
		Header: &proto.EventHeader{
			DeviceCode: deviceCode,
			Version:    1,
			Date:       t,
			EventType:  proto.EventHeader_EventType(rand.Intn(5) + 1),
		},
		Body: &proto.EventBody{
			Files: []*proto.File{
				{
					Time:     t,
					Category: rand.Int31n(5) + 1,
					Data:     images[rand.Intn(3)],
				},
			},
		},
	}
}
