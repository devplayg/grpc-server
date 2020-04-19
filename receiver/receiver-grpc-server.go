package receiver

import (
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"net"
)

func (r *Receiver) getGrpcServerOptions() []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)

	// Set statsHandler
	opts = append(opts, grpc.StatsHandler(&grpc_server.ConnStatsHandler{
		From: "agent",
		Log:  r.Log,
	}))

	// Keepalive Enforcement policy
	// https://godoc.org/google.golang.org/grpc/keepalive#EnforcementPolicy
	opts = append(opts, grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		// MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		// PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}))

	// Keepalive server parameters
	// https://godoc.org/google.golang.org/grpc/keepalive#ServerParameters
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		// MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		// MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		// MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		// Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		// Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
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
		r.Log.Infof("secured gRPC with %s", creds.Info().SecurityProtocol)
	}
	// grpc.UnaryInterceptor(grpc_server.UnaryInterceptor),

	return opts
}

func (r *Receiver) startGrpcServer(gRpcClient proto.EventServiceClient) error {
	ln, err := net.Listen("tcp", r.config.App.Receiver.Address)
	if err != nil {
		return err
	}
	r.Log.Infof("gRPC server is listening on %s for requests from agent", r.config.App.Receiver.Address)

	opts := r.getGrpcServerOptions()
	r.gRpcServer = grpc.NewServer(opts...)

	// Register server to gRPC server
	proto.RegisterEventServiceServer(r.gRpcServer, &grpcService{target: gRpcClient, Log: r.Log})

	// Run
	if err := r.gRpcServer.Serve(ln); err != nil {
		return err
	}
	return nil
}
