package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/devplayg/hippo/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"runtime"
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
	workerCount := runtime.NumCPU() * 2
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

	// Handle TX failed events
	wg.Add(1)
	go func() {
		log.Debug("tx-failed-handler has been started")
		defer func() {
			log.Debug("tx-failed-handler has been stopped")
			wg.Done()
		}()
		if err := r.handleTxFailedEvent(); err != nil {
			log.Error(fmt.Errorf("failed to run tx-failed-handler: %w", err))
			r.Cancel()
			return
		}
	}()

	// Run gRPC server
	wg.Add(1)
	go func() {
		log.WithFields(logrus.Fields{
			"secured": !r.Engine.Config.Insecure,
			"address": r.config.App.Receiver.Address,
		}).Debug("gRPC service has been started")
		defer func() {
			log.Debug("gRPC service has been stopped")
			wg.Done()
		}()

		if err := r.startGrpcServer(r.storageCh); err != nil {
			log.Error(fmt.Errorf("failed to start gRPC server: %w", err))
			r.Cancel()
			return
		}
	}()
	log.WithFields(logrus.Fields{
		"batchSize":        r.batchSize,
		"batchTimeout(ms)": r.batchTimeout.Milliseconds(),
		"workerCount":      r.workerCount,
	}).Infof("server(%s) has been started", r.Engine.Config.Name)

	// Run monitoring service
	if r.monitor {
		wg.Add(1)
		go func() {
			log.WithFields(logrus.Fields{
				"address": r.monitorAddr,
			}).Debug("monitoring service has been started")
			defer func() {
				log.Debug("monitoring service has been stopped")
				wg.Done()
			}()

			if err := r.startMonitor(); err != nil {
				log.Error(err)
				r.Cancel()
				return
			}
		}()
	}

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
