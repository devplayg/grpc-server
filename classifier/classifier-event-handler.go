package classifier

import (
	"bytes"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
	"sync"
	"time"
)

func (c *Classifier) save(event *proto.Event) error {
	deviceId, _ := c.deviceCodeMap[event.Header.DeviceCode]
	eventWrapper := &EventWrapper{
		event:    event,
		uuid:     uuid.New(),
		deviceId: deviceId,
	}

	wg := new(sync.WaitGroup)

	// Save header
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.saveHeader(eventWrapper); err != nil {
			log.Error(fmt.Errorf("failed to insert; %w", err))
			return
		}
	}()

	// Save body
	wg.Add(1)
	go func() {
		defer func() {
			log.Tracef("uploading %s [done]", eventWrapper.uuid)
			wg.Done()
		}()
		log.Tracef("uploading %s", eventWrapper.uuid)
		if err := c.saveBody(eventWrapper); err != nil {
			log.Error(fmt.Errorf("failed to insert; %w", err))
			return
		}
	}()

	// Wait
	wg.Wait()

	return nil
}

func (c *Classifier) saveHeader(event *EventWrapper) error {
	started := time.Now()
	db := c.db.Create(event.entity())
	if db.Error != nil {
		return db.Error
	}
	dur := time.Since(started)
	grpc_server.ServerStats.Add(statsInsertingTime, dur.Milliseconds())
	return nil
}

func (c *Classifier) saveBody(e *EventWrapper) error {
	var total int64

	r := bytes.NewReader(nil)
	for i, f := range e.event.Body.Files {
		size := int64(len(f.Data))
		r.Reset(f.Data)

		started := time.Now()
		_, err := c.minioClient.PutObject(
			c.config.App.Storage.Bucket,
			fmt.Sprintf("%s_%d.jpg", e.uuid.String(), i),
			r,
			size,
			minio.PutObjectOptions{ContentType: ": image/jpeg"},
		)
		if err != nil {
			log.Error(err)
			continue
		}

		grpc_server.ServerStats.Add(grpc_server.StatsCount, 1)
		grpc_server.ServerStats.Add(grpc_server.StatsSize, size)
		grpc_server.ServerStats.Add(grpc_server.StatsWorkingTime, time.Since(started).Milliseconds())
		total += size
	}
	return nil
}
