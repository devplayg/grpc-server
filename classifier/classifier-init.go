package classifier

import (
	"expvar"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/minio/minio-go"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Classifier) init() error {
	// Initialize logger
	log = c.Log
	stats = expvar.NewMap(c.Engine.Config.Name)

	// Initialize configuration
	config, err := grpc_server.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration")
	}
	if len(config.App.Classifier.Address) < 1 {
		config.App.Classifier.Address = "127.0.0.1:8802"
	}
	c.config = config

	// Initialize database
	if err := c.initDatabase(4, 4); err != nil {
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
	//minioClient.MakeBucket(c.config.App.Storage.Bucket, "")
	c.minioClient = minioClient

	// Load asset
	if err := c.loadDevices(); err != nil {
		return fmt.Errorf("failed to load devices; %w", err)
	}

	// Monitoring
	if err := c.initMonitor(); err != nil {
		return err
	}

	return nil
}

func (c *Classifier) loadDevices() error {
	c.deviceCodeMap = make(map[string]int64)
	start := 65
	count := 26
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

func (c *Classifier) initMonitor() error {
	stats.Set("start", new(expvar.Int))
	stats.Set("end", new(expvar.Int))
	stats.Set("inserted-time", new(expvar.Int))
	stats.Set("uploaded-time", new(expvar.Int))

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		m := map[string]interface{}{
			"duration":      (stats.Get("end").(*expvar.Int).Value() - stats.Get("start").(*expvar.Int).Value()) / int64(time.Millisecond),
			"inserted-time": stats.Get("inserted-time").(*expvar.Int).Value(),
			"uploaded-time": stats.Get("uploaded-time").(*expvar.Int).Value(),
		}

		s := fmt.Sprintf("%d\t%d\t%d",
			m["duration"],
			m["inserted-time"],
			m["uploaded-time"],
		)
		//dur := stats.Get("end").(*expvar.Int).Value() - stats.Get("start").(*expvar.Int).Value()
		//m["relayed-time"] = dur / int64(time.Millisecond)
		//m["relayed"] = stats.Get("relayed").(*expvar.Int).Value()
		//b, _ := json.MarshalIndent(m, "", "  ")
		w.Write([]byte(s))
	})

	go http.ListenAndServe(":8123", nil)

	return nil
}
