package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	fs         = pflag.NewFlagSet("generator", pflag.ExitOnError)
	agentCount = fs.IntP("agent", "a", 10, "Client count")
	dataCount  = fs.IntP("c", "c", 1, "Event count by client")
	addr       = fs.String("addr", "localhost:8801", "Receiver address")
	insecure   = fs.Bool("insecure", false, "Disable TLS")
	devices    []string
	images     [][]byte

	conn      *grpc.ClientConn
	clientApi proto.EventServiceClient

	size uint64
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

func main() {
	// Generate sample data
	data := generateData()

	// Connect to receiver
	conn, clientApi = connect()
	defer conn.Close()

	// Reset debug
	clientApi.ResetDebug(context.Background(), &empty.Empty{})

	// Run
	wg := new(sync.WaitGroup)
	started := time.Now()
	for i := 0; i < *agentCount; i++ {
		wg.Add(1)
		k := int32(i)
		go send(wg, data[k])
	}
	wg.Wait()

	dur := time.Since(started)
	fmt.Printf("agent=%d, totalData=%d, time=%d\n", *agentCount, (*agentCount)*(*dataCount), dur.Milliseconds())

	startHttpServer(dur)
	fmt.Scanln()
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

func connect() (*grpc.ClientConn, proto.EventServiceClient) {

	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{}))
	if !*insecure {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Create connection
	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		panic(err)
	}

	// Create client API for service
	clientApi := proto.NewEventServiceClient(conn)
	return conn, clientApi

}

func startHttpServer(dur time.Duration) {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := clientApi.GetServerStats(context.Background(), &empty.Empty{})
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		s := fmt.Sprintf("%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%s",
			*agentCount,
			*dataCount,

			// Receiver
			(*agentCount)*(*dataCount),
			dur.Milliseconds(),
			size,

			// Classifier stats
			stats.Count,
			(stats.EndTimeUnixNano-stats.StartTimeUnixNano)/int64(time.Millisecond),
			stats.Size,

			stats.Meta,
		)
		w.Write([]byte(s))
	})

	go http.ListenAndServe("127.0.0.1:8120", nil)
}

func send(wg *sync.WaitGroup, events []*proto.Event) {
	defer wg.Done()

	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			// Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			// Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			// PermitWithoutStream: true,             // send pings even without active streams
		}),
	}
	if !*insecure {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Create connection
	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create client API for service
	clientApi := proto.NewEventServiceClient(conn)

	// gRPC remote procedure call
	for _, e := range events {
		for _, f := range e.Body.Files {
			atomic.AddUint64(&size, uint64(len(f.Data)))
		}
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
					Data:     images[0],
				},
			},
		},
	}
}
