package classifier

import (
	"context"
	"fmt"
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
	"sync"
)

func (c *Classifier) startGrpcService(wg *sync.WaitGroup, serviceName string) {
	go func() {
		log.WithFields(logrus.Fields{
			"secured": !c.Engine.Config.Insecure,
			"address": c.config.App.Classifier.Address,
			"worker":  c.workerCount,
		}).Debugf("%s has been started", serviceName)

		defer func() {
			log.Debugf("%s has been stopped", serviceName)
			wg.Done()
		}()

		if err := c._startGrpcServer(); err != nil {
			log.Error(fmt.Errorf("failed to start %s: %w", serviceName, err))
			c.Cancel()
			return
		}
	}()
}

func (c *Classifier) _startGrpcServer() error {
	ln, err := net.Listen("tcp", c.config.App.Classifier.Address)
	if err != nil {
		return err
	}

	opts := c.getGrpcServerOptions()
	c.gRpcServer = grpc.NewServer(opts...)

	// Register server to gRPC server
	service := &grpcService{
		classifier:     c,
		notifierClient: c.notifierClient,
		ch:             make(chan bool, c.workerCount),
	}
	proto.RegisterEventServiceServer(c.gRpcServer, service)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	// Run gRPC server
	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := c.gRpcServer.Serve(ln); err != nil {
			ctx = context.WithValue(ctx, "err", err)
			cancel()
			return
		}
	}()

	// Wait for signal
	select {
	case <-ctx.Done(): // from local context
		<-ch
		return ctx.Value("err").(error)

	case <-c.Ctx.Done(): // from server context
		log.Debug(fmt.Errorf("gRPC service received stop signal"))
		c.gRpcServer.Stop()
		<-ch
	}

	return nil
}

func (c *Classifier) getGrpcServerOptions() []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)

	// Set statsHandler
	opts = append(opts, grpc.StatsHandler(&grpc_server.ConnStatsHandler{
		From: "receiver",
		Log:  c.Log,
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
	if c.Engine.Config.Trace {
		logrusEntry := logrus.NewEntry(c.Log)
		opts = append(opts, grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(
				grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor),
			),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
		))
	}

	// Set security
	if !c.Engine.Config.Insecure {
		// Create gRPC server
		creds, err := credentials.NewServerTLSFromFile(c.Engine.Config.CertFile, c.Engine.Config.KeyFile)
		if err != nil {
			panic(err)
		}
		opts = append(opts, grpc.Creds(creds))
		log.Infof("secured gRPC with %s", creds.Info().SecurityProtocol)
	}
	// grpc.UnaryInterceptor(grpc_server.UnaryInterceptor),

	return opts
}
