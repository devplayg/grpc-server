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
	"strings"
	"sync"

	"time"
)

var (
	fs         = pflag.NewFlagSet("generator", pflag.ExitOnError)
	agentCount = fs.IntP("agent", "a", 10, "Client count")
	dataCount  = fs.IntP("c", "c", 1, "Event count by client")
	receiver   = fs.String("recv", "localhost:8801", "Receiver address")
	classifier = fs.String("classy", "localhost:8802", "Classifier address")
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
	wg := new(sync.WaitGroup)

	// Reset
	reset()

	started := time.Now()
	for i := 0; i < *agentCount; i++ {
		wg.Add(1)
		k := int32(i)
		go send(wg, data[k])
	}

	wg.Wait()
	dur := time.Since(started)
	fmt.Printf("agent=%d, totalData=%d, time=%d\n", *agentCount, (*agentCount)*(*dataCount), dur.Milliseconds())

	startHttpServer(*agentCount, *dataCount, dur)
	fmt.Scanln()
}

func reset() {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{}),
	}

	// Create connection
	conn, err := grpc.Dial(*receiver, opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create client API for service
	clientApi := proto.NewEventServiceClient(conn)

	// gRPC remote procedure call
	_, err = clientApi.ResetStats(context.Background(), &empty.Empty{})
	if err != nil {
		panic(err)
	}
}

func startHttpServer(agentCount, dataCountByAgent int, dur time.Duration) {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		arr1 := strings.Split(*receiver, ":")
		arr2 := strings.Split(*classifier, ":")
		receiverUrl := fmt.Sprintf("http://%s:8123/stats", arr1[0])
		classifierUrl := fmt.Sprintf("http://%s:8124/stats", arr2[0])

		s := fmt.Sprintf("%d\t%d\t%d\t%d\t%s\t%s",
			agentCount,
			dataCountByAgent,
			agentCount*dataCountByAgent,
			dur.Milliseconds(),
			getRemoteStats(receiverUrl),
			getRemoteStats(classifierUrl),
		)
		w.Write([]byte(s))
	})

	go http.ListenAndServe("127.0.0.1:8120", nil)
}

func getRemoteStats(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(data)
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
	conn, err := grpc.Dial(*receiver, opts...)
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
					Data:     images[0],
				},
			},
		},
	}
}
