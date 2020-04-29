package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"time"
)

func (c *Classifier) init() error {
	log = c.Log
	grpc_server.ResetServerStats(statsInsertingTime)

	if err := c.loadConfig(); err != nil {
		return err
	}

	// Initialize database
	if err := c.initDatabase(c.workerCount, c.workerCount); err != nil {
		return fmt.Errorf("failed to initialize database; %w", err)
	}

	// Initialize MinIO
	minioClient, err := minio.New(
		c.config.App.Storage.Address,
		c.config.App.Storage.AccessKey,
		c.config.App.Storage.SecretKey,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO; %w", err)
	}

	log.WithFields(logrus.Fields{
		"address":       c.config.App.Storage.Address,
		"retention-sec": c.dataRetentionTime.Seconds(),
	}).Debug("object storage")
	c.minioClient = minioClient

	// Load asset
	if err := c.loadDevices(); err != nil {
		return fmt.Errorf("failed to load devices; %w", err)
	}

	return nil
}

func (c *Classifier) loadConfig() error {
	config, err := grpc_server.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration")
	}
	if len(config.App.Classifier.Address) < 1 {
		config.App.Classifier.Address = DefaultAddress
	}

	if config.App.Storage.RetentionTimeSec < 1 {
		config.App.Storage.RetentionTimeSec = 24 * 60 * 60
	}
	c.dataRetentionTime = time.Duration(config.App.Storage.RetentionTimeSec) * time.Second
	c.config = config
	return nil
}

func (c *Classifier) loadDevices() error {
	c.deviceCodeMap = make(map[string]int64)
	start := 65 // From 'A'
	count := 26 // To 'Z'
	id := int64(1)
	for i := start; i < start+count; i++ {
		code := string(i)
		c.deviceCodeMap["DEVICE-"+code] = id
		id++
	}

	return nil
}

func (c *Classifier) initDatabase(maxIdleConns, maxOpenConns int) error {
	connStr, dbTimezone, err := convertJdbcUrlToGoOrm(c.config.Spring.Datasource.Url, c.config.Spring.Datasource.Username, c.config.Spring.Datasource.Password)
	if err != nil {
		return err
	}
	log.Tracef("database connection: %s", connStr)

	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		return err
	}
	db.DB().SetMaxIdleConns(maxIdleConns)
	db.DB().SetMaxOpenConns(maxOpenConns)
	c.db = db
	c.dbTimezone = dbTimezone
	log.WithFields(logrus.Fields{
		"maxIdleConns": maxIdleConns,
		"maxOpenConns": maxOpenConns,
	}).Debug("database")

	return nil
}
