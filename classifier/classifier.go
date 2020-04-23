package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/hippo/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"runtime"
	"time"
)

var (
	log *logrus.Logger
)

const (
	DefaultAddress = "127.0.0.1:8802"
)

// Classifier receives data from receiver via gRPC framework
type Classifier struct {
	hippo.Launcher
	config         *grpc_server.Config
	gRpcServer     *grpc.Server
	gRpcClientConn *grpc.ClientConn
	notifier       *notifier
	batchSize      int
	batchTimeout   time.Duration
	storage        string
	workerCount    int

	// Database
	db         *gorm.DB
	dbTimezone *time.Location

	// Device
	deviceCodeMap map[string]int64

	// Storage
	minioClient *minio.Client
}

func NewClassifier(batchSize int, batchTimeout time.Duration, worker int) *Classifier {
	workerCount := runtime.NumCPU() * 2
	if worker > 0 {
		workerCount = worker
	}

	return &Classifier{
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		workerCount:  workerCount,
	}
}

func (c *Classifier) Start() error {
	if err := c.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", c.Engine.Config.Name, err)
	}

	// Connect to notifier
	c.notifier = newNotifier(c.config.App.Classifier.Notifier.Address)
	if err := c.notifier.connect(); err != nil {
		return fmt.Errorf("failed to connect to notifier: %w", err)
	}

	// Start event handler
	//if err := c.handleEvent(); err != nil {
	//	return fmt.Errorf("failed to start event handler; %w", err)
	//}

	ch := make(chan bool)
	go func() {
		defer close(ch)
		if err := c.startGrpcServer(); err != nil {
			log.Error(fmt.Errorf("failed to start gRPC server: %w", err))
			c.Cancel()
			return
		}
		log.Debug("gRpcServer has been stopped")
	}()
	log.WithFields(logrus.Fields{
		"batchSize":        c.batchSize,
		"batchTimeout(ms)": c.batchTimeout.Milliseconds(),
		"workerCount":      c.workerCount,
	}).Infof("%s has been started", c.Engine.Config.Name)

	<-c.Ctx.Done()

	// Stop gRPC server
	if c.gRpcServer != nil {
		c.gRpcServer.Stop()
	}

	// Waiting for gRPC server to shut down
	<-ch
	return nil
}

func (c *Classifier) Stop() error {
	defer c.Log.Infof("%s has been stopped", c.Engine.Config.Name)

	if c.notifier != nil {
		if err := c.notifier.disconnect(); err != nil {
			c.Log.Error("failed to disconnect classifier")
		}
	}

	return nil
}
