package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/minio/minio-go"
	"net/url"
	"strings"
	"time"
)

func (c *Classifier) init() error {
	log = c.Log
	grpc_server.ResetServerStats(statsInsertingTime)

	if err := c.loadConfig(); err != nil {
		return err
	}

	// Initialize database
	if err := c.initDatabase(c.workerCount+4, c.workerCount+4); err != nil {
		return fmt.Errorf("failed to initialize database; %w", err)
	}

	// Initialize MinIO
	//spew.Dump(c.config.App.Storage)
	minioClient, err := minio.New(
		c.config.App.Storage.Address,
		c.config.App.Storage.AccessKey,
		c.config.App.Storage.SecretKey,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO; %w", err)
	}

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
	connStr, dbTimezone, err := ConvertJdbcUrlToGoOrm(c.config.Spring.Datasource.Url, c.config.Spring.Datasource.Username, c.config.Spring.Datasource.Password)
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

	return nil
}

func ConvertJdbcUrlToGoOrm(str, username, password string) (string, *time.Location, error) {
	jdbcUrl := strings.TrimPrefix(str, "jdbc:")
	u, err := url.Parse(jdbcUrl)
	if err != nil {
		return "", nil, err
	}

	param, err := url.ParseQuery(u.RawQuery)
	timezone := param.Get("serverTimezone")
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", nil, err
	}

	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)%s?allowAllFiles=true&charset=utf8&parseTime=true&loc=%s",
		username,
		password,
		u.Host,
		u.Path,
		strings.Replace(param.Get("serverTimezone"), "/", "%2F", -1),
	)

	return connStr, loc, nil
}
