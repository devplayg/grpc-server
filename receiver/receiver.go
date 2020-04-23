package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/devplayg/hippo/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"runtime"
	"time"
)

const (
	DefaultStorageDir = "data"
	DefaultAddress    = "127.0.0.1:8801"
)

var (
	log *logrus.Logger
)

// Receiver receives data from agents via gRPC framework
type Receiver struct {
	hippo.Launcher
	config           *grpc_server.Config
	gRpcServer       *grpc.Server
	classifierClient *classifierClient
	batchSize        int
	batchTimeout     time.Duration
	storage          string
	storageCh        chan *proto.Event
	workerCount      int
}

func NewReceiver(batchSize int, batchTimeout time.Duration, worker int) *Receiver {
	workerCount := runtime.NumCPU() * 2
	if worker > 0 {
		workerCount = worker
	}

	return &Receiver{
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		storageCh:    make(chan *proto.Event, batchSize),
		workerCount:  workerCount,
	}
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}

	// Classifier
	r.classifierClient = newClassifierClient(r.config.App.Receiver.Classifier.Address, r.Engine.Config.Insecure)
	log.Debugf("connecting to classifier %s", r.config.App.Receiver.Classifier.Address)
	if err := r.classifierClient.connect(); err != nil {
		return fmt.Errorf("failed to connect to classifier: %w", err)
	}

	// Handle TX failed events
	if err := r.handleTxFailedEvent(); err != nil {
		return fmt.Errorf("failed to run storage channal: %w", err)
	}

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := r.startGrpcServer(r.storageCh); err != nil {
			log.Error(fmt.Errorf("failed to start gRPC server: %w", err))
			r.Cancel()
			return
		}
		log.Debug("gRpcServer has been stopped")
	}()
	log.WithFields(logrus.Fields{
		"batchSize":        r.batchSize,
		"batchTimeout(ms)": r.batchTimeout.Milliseconds(),
		"workerCount":      r.workerCount,
	}).Infof("%s has been started", r.Engine.Config.Name)

	<-r.Ctx.Done()

	// Stop gRPC server
	if r.gRpcServer != nil {
		r.gRpcServer.Stop()
	}

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (r *Receiver) Stop() error {
	defer log.Infof("%s has been stopped", r.Engine.Config.Name)

	if r.classifierClient != nil {
		if err := r.classifierClient.disconnect(); err != nil {
			log.Error("failed to disconnect classifier")
		}
	}

	return nil
}
