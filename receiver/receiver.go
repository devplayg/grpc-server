package receiver

import (
	"context"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/devplayg/hippo/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"log"
	"net"
	"reflect"
	"time"
)

type Receiver struct {
	hippo.Launcher
	config       *grpc_server.Config
	grpcReceiver *grpc.Server
	//grpcSender   proto.EventServiceClient
}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}
	r.Log.Infof("%s has been started", r.Engine.Config.Name)

	grpcSender, err := r.startGRPCSender()
	if err != nil {
		return fmt.Errorf("failed to start gRPC sender")
	}

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := r.startGRPCReceiver(grpcSender); err != nil {
			r.Log.Errorf("failed to start gRPC server: %w", err)
			return
		}
		r.Log.Debug("gRpcServer has been stopped")
	}()

	<-r.Ctx.Done()

	// Stop gRPC server
	r.grpcReceiver.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	r.Log.Infof("%s has been stopped", r.Engine.Config.Name)
	return nil
}

func (r *Receiver) startGRPCReceiver(grpcSender proto.EventServiceClient) error {
	ln, err := net.Listen("tcp", r.config.App.Receiver.Address)
	if err != nil {
		return err
	}

	//logrusEntry := logrus.NewEntry(r.Log)
	//opts := []grpc_logrus.Option{
	//	grpc_logrus.WithLevels(customFunc),
	//}
	// Make sure that log statements internal to gRPC library are logged using the logrus Logger as well.
	//grpc_logrus.ReplaceGrpcLogger(logrusEntry)

	//logrus.ErrorKey = "grpc.error"
	//logrusEntry := logrus.NewEntry(r.Log)

	// Create gRPC server

	r.grpcReceiver = grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StatsHandler(&ServerHandler{}),
		//grpc_logrus.UnaryServerInterceptor(logrusEntry),
		//grpc_middleware.WithUnaryServerChain(
		//	grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		//	grpc_logrus.UnaryServerInterceptor(logrusEntry, opts...),
		//),
	)

	// Register server to gRPC server
	proto.RegisterEventServiceServer(r.grpcReceiver, &eventReceiver{grpcSender})

	// Run
	//r.Log.Debug("gRpcServer has been started")
	if err := r.grpcReceiver.Serve(ln); err != nil {
		return err
	}
	return nil
}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	//
	//skip auth when ListReleases requested
	//if info.FullMethod != "/proto.GoReleaseService/ListReleases" {
	//	if err := authorize(ctx); err != nil {
	//		return nil, err
	//	}
	//}
	h, err := handler(ctx, req)

	//logging
	log.Printf("request - Method:%s\tDuration:%s\tError:%v\n",
		info.FullMethod,
		time.Since(start),
		err)

	return h, err
}

func (r *Receiver) startGRPCSender() (proto.EventServiceClient, error) {
	conn, err := grpc.Dial(r.config.App.Receiver.Classifier.Address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create client API for service
	grpcSender := proto.NewEventServiceClient(conn)

	// gRPC remote procedure call
	//for {
	//	event := generateEvent()
	//	spew.Dump(event)
	//	_, err := clientApi.Send(context.Background(), event)
	//	if err != nil {
	//		panic(err)
	//	}
	//	time.Sleep(100 * time.Millisecond)
	//
	//}
	return grpcSender, nil
}

type ServerHandler struct {
}

func (s *ServerHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	//fmt.Printf("TagRPC\tinfo.FullMethodName=%s\tinfo.FailFast=%v\n", info.FullMethodName, info.FailFast)
	return ctx
}
func (s *ServerHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	//fmt.Printf("HandleRPC\trpcStats.IsClient()=%v\n", rpcStats.IsClient())
}
func (s *ServerHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	ctx = context.WithValue(ctx, "remoteaddr", info.RemoteAddr.String())
	fmt.Printf("TagConn\tinfo.RemoteAddr=%v\n", info.RemoteAddr)
	return ctx
}
func (s *ServerHandler) HandleConn(ctx context.Context, connStats stats.ConnStats) {
	//a := reflect.TypeOf(connStats)
	//spew.Dump(a)
	//spew.Dump(ctx)
	if reflect.TypeOf(connStats).String() == "*stats.ConnBegin" {
		fmt.Printf("HandleConn\t connected=%v\n", connStats.IsClient())
		return
	}
	fmt.Printf("HandleConn\t disconnected=%v\n", connStats.IsClient())

	if v := ctx.Value("remoteaddr"); v != nil {
		// 타입 확인(type assertion)
		u, ok := v.(string)
		if ok {
			fmt.Printf("HandleConn\t disconnected=%s\n", u)
		}
	}

	//spew.Dump(a)
	//a := connStats.(stats.ConnBegin)
	//reflect.TypeOf(connStats).

	//if reflect.TypeOf(connStats) == stats.ConnEnd.(type) { {

	//}
	//r1 := reflect.TypeOf(connStats)
	//r2 := reflect.TypeOf(stats.ConnBegin)
	// 호출된 Controller 접근권한 덮어쓰기
	//if app, ok := c.AppController.(CtrlPreparer); ok {
	//	app.CtrlPrepare()
	//}

}
