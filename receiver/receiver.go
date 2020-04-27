package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/devplayg/hippo/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"sync"
	"time"
)

const (
	DefaultStorageDir     = "data"
	DefaultAddress        = ":8801"
	DefaultMonitorAddress = ":8901"
)

var (
	log *logrus.Logger
)

func NewReceiver(batchSize int, batchTimeout time.Duration, worker int, monitor bool, monitorAddr string) *Receiver {
	workerCount := 4000
	if worker > 0 {
		workerCount = worker
	}
	if len(monitorAddr) < 1 {
		monitorAddr = DefaultMonitorAddress
	}

	return &Receiver{
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		storageCh:    make(chan *proto.Event, batchSize),
		workerCount:  workerCount,
		monitor:      monitor,
		monitorAddr:  monitorAddr,
	}
}

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
	monitor          bool
	monitorAddr      string
}

func (r *Receiver) Start() error {
	if err := r.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", r.Engine.Config.Name, err)
	}

	// Connect to classifier
	r.classifierClient = newClassifierClient(r.config.App.Receiver.Classifier.Address, r.Engine.Config.Insecure)
	if err := r.classifierClient.connect(); err != nil {
		return fmt.Errorf("failed to connect to classifier: %w", err)
	}

	// Create wait group
	wg := new(sync.WaitGroup)

	// Start tx-failed-handler
	wg.Add(1)
	r.startTxHandler(wg, "txHandler")

	// Start gRPC service
	wg.Add(1)
	r.startGrpcService(wg, "gRPC service")

	// Start  monitoring service
	if r.monitor {
		wg.Add(1)
		r.startMonitoringService(wg, "monitoring service")
	}

	log.Infof("server(%s) has been started", r.Engine.Config.Name)

	// Wait for canceling context
	<-r.Ctx.Done()

	// Waiting for gRPC server to shut down
	wg.Wait()

	return nil
}

func (r *Receiver) Stop() error {
	defer log.Infof("server(%s) has been stopped", r.Engine.Config.Name)

	if r.classifierClient != nil {
		if err := r.classifierClient.disconnect(); err != nil {
			log.Error("failed to disconnect classifier")
		}
	}

	return nil
}
