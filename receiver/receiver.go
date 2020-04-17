package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/devplayg/hippo/v2"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

// Receiver receives data from agents via gRPC framework
type Receiver struct {
	hippo.Launcher
	config         *grpc_server.Config
	recv           *grpc.Server     // gRPC receiver
	classifierConn *grpc.ClientConn // gRPC client for classifier
}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (r *Receiver) Start() error {
	// Initialize receiver
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}
	r.Log.Debugf("%s has been initialized", r.Engine.Config.Name)

	// Connect with classifier
	grpcOutput, err := r.connectWithClassifier()
	if err != nil {
		return fmt.Errorf("failed to start gRPC sender")
	}

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := r.startGrpcServer(grpcOutput); err != nil {
			r.Log.Errorf("failed to start gRPC server: %w", err)
			return
		}
		r.Log.Debug("gRpcServer has been stopped")
	}()

	<-r.Ctx.Done()

	// Stop gRPC server
	r.recv.Stop()

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	r.Log.Infof("%s has been stopped", r.Engine.Config.Name)
	if r.classifierConn == nil {
		r.Log.Debugf("connection status=%v", r.classifierConn.GetState())
		if err := r.classifierConn.Close(); err != nil {
			r.Log.Error(err)
		}
	}

	return nil
}

func (r *Receiver) startGrpcServer(grpcSender proto.EventServiceClient) error {
	ln, err := net.Listen("tcp", r.config.App.Receiver.Address)
	if err != nil {
		return err
	}
	r.Log.Infof("%s is listening on %s", r.Engine.Config.Name, r.config.App.Receiver.Address)

	// Set options
	opts := make([]grpc.ServerOption, 0)

	// Set statsHandler

	opts = append(opts, grpc.StatsHandler(&grpc_server.ConnStatsHandler{
		From: "agent",
		To:   fmt.Sprintf("%s(%s)", r.Engine.Config.Name, r.config.App.Receiver.Address),
		Log:  r.Log,
	}))

	// Set logging
	if r.Engine.Config.Debug {
		logrusEntry := logrus.NewEntry(r.Log)
		opts = append(opts, grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(
				grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor),
			),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
		))
	}

	// Set security
	if !r.Engine.Config.Insecure {
		// Create gRPC server
		creds, err := credentials.NewServerTLSFromFile(r.Engine.Config.CertFile, r.Engine.Config.KeyFile)
		if err != nil {
			panic(err)
		}
		opts = append(opts, grpc.Creds(creds))
		r.Log.Infof("secured %s", creds.Info().SecurityProtocol)
	}
	//grpc.UnaryInterceptor(grpc_server.UnaryInterceptor),

	r.recv = grpc.NewServer(opts...)

	// Register server to gRPC server
	proto.RegisterEventServiceServer(r.recv, &output{grpcSender: grpcSender, Log: r.Log})

	// Run
	if err := r.recv.Serve(ln); err != nil {
		return err
	}
	return nil
}

func (r *Receiver) connectWithClassifier() (proto.EventServiceClient, error) {
	classifierConn, err := grpc.Dial(
		r.config.App.Receiver.Classifier.Address,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&grpc_server.ConnStatsHandler{
			From: r.Engine.Config.Name,
			To:   fmt.Sprintf("%s(%s)", "classifier", r.config.App.Receiver.Classifier.Address),
			Log:  r.Log,
		}),
	)
	if err != nil {
		return nil, err
	}
	r.classifierConn = classifierConn

	// Create client API for service
	classifierApi := proto.NewEventServiceClient(r.classifierConn)
	return classifierApi, nil
}
