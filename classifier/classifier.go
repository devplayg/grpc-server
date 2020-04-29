package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/hippo/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var (
	log *logrus.Logger
)

const (
	DefaultAddress        = ":8802"
	DefaultMonitorAddress = ":8902"
)

func NewClassifier(batchSize int, batchTimeout time.Duration, worker int, monitor bool, monitorAddr string) *Classifier {
	workerCount := 4000
	if worker > 0 {
		workerCount = worker
	}
	if len(monitorAddr) < 1 {
		monitorAddr = DefaultMonitorAddress
	}

	return &Classifier{
		batchSize:     batchSize,
		batchTimeout:  batchTimeout,
		workerCount:   workerCount,
		monitor:       monitor,
		monitorAddr:   monitorAddr,
		eventHeaderCh: make(chan *EventWrapper, workerCount),
		eventBodyCh:   make(chan bool, workerCount),
	}
}

// Classifier receives data from receiver via gRPC framework
type Classifier struct {
	hippo.Launcher
	config         *grpc_server.Config
	gRpcServer     *grpc.Server
	gRpcClientConn *grpc.ClientConn
	notifierClient *notifierClient
	batchSize      int
	batchTimeout   time.Duration
	storage        string
	workerCount    int
	monitor        bool
	monitorAddr    string

	// Database
	db         *gorm.DB
	dbTimezone *time.Location

	// Device
	deviceCodeMap map[string]int64

	// Storage
	minioClient *minio.Client

	// Event channel
	eventHeaderCh chan *EventWrapper
	eventBodyCh   chan bool

	dataRetentionTime time.Duration
}

func (c *Classifier) Start() error {
	if err := c.init(); err != nil {
		return fmt.Errorf("failed to initialize %s; %w", c.Engine.Config.Name, err)
	}

	// Connect to notifier
	c.notifierClient = newNotifier(c.config.App.Classifier.Notifier.Address)
	if err := c.notifierClient.connect(); err != nil {
		return fmt.Errorf("failed to connect to notifier: %w", err)
	}

	// Create wait group
	wg := new(sync.WaitGroup)

	wg.Add(1)
	c.saveHeader(wg)

	// Start gRPC service
	wg.Add(1)
	c.startGrpcService(wg, "gRPC service")

	// Start  monitoring service
	if c.monitor {
		wg.Add(1)
		c.startMonitoringService(wg, "monitoring service")
	}

	log.Infof("server(%s) has been started", c.Engine.Config.Name)

	// Wait for canceling context
	<-c.Ctx.Done()

	// Waiting for gRPC server to shut down
	wg.Wait()

	return nil
}

func (c *Classifier) Stop() error {
	defer c.Log.Infof("%s has been stopped", c.Engine.Config.Name)

	if c.notifierClient != nil {
		if err := c.notifierClient.disconnect(); err != nil {
			c.Log.Error("failed to disconnect classifier")
		}
	}

	return nil
}
