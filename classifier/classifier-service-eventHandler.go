package classifier

import (
	"bytes"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v6"
	"os"
	"sync"
	"time"
)

func (c *Classifier) wrapEvent(event *proto.Event) *EventWrapper {
	deviceId, _ := c.deviceCodeMap[event.Header.DeviceCode]
	return &EventWrapper{
		event:    event,
		uuid:     uuid.New(),
		deviceId: deviceId,
	}
}

func (c *Classifier) save(e *proto.Event) error {
	eventWrapper := c.wrapEvent(e)

	// Save event header
	c.eventHeaderCh <- eventWrapper

	// Save event body
	c.eventBodyCh <- true
	go func() {
		defer func() {
			<-c.eventBodyCh
		}()

		if err := c.saveBody(eventWrapper); err != nil {
			log.Error(err)
			return
		}
	}()
	return nil
}

func (c *Classifier) saveHeader(wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()

		batch := make([]*EventWrapper, 0, c.batchSize)
		timer := time.NewTimer(c.batchTimeout)
		timer.Stop()

		save := func() {
			started := time.Now()
			defer func() {
				batch = make([]*EventWrapper, 0, c.batchSize)
				grpc_server.ServerStats.Add(statsInsertingTime, time.Since(started).Milliseconds())
			}()
			path, err := writeTextIntoTempFile(grpc_server.TempDir, eventsToTsv(batch))
			if err != nil {
				log.Error(err)
				return
			}
			db := grpc_server.BulkInsert(c.db, "log", "date, device_id, event_type, UUID, flag", path)
			if db.Error != nil {
				log.Error(err)
				return
			}
			log.Debugf("inserted %d row(s)", len(batch))
			os.Remove(path)
		}

		for {
			select {
			case eventWrapper := <-c.eventHeaderCh:
				batch = append(batch, eventWrapper)

				if len(batch) == 1 {
					timer.Reset(c.batchTimeout)
				}
				if len(batch) == c.batchSize {
					timer.Stop()
					save()
				}
			case <-c.Ctx.Done():
				//log.Debug("storage channel is done")
				if len(batch) > 0 {
					save()
					return
				}
				return
			case <-timer.C:
				save()
			}
		}
	}()

	return nil
}

func (c *Classifier) saveBody(e *EventWrapper) error {
	r := bytes.NewReader(nil)
	for i, f := range e.event.Body.Files {
		size := int64(len(f.Data))
		r.Reset(f.Data)

		started := time.Now()
		object := fmt.Sprintf("%s_%d.jpg", e.uuid.String(), i)
		//t := time.Now().Add(c.dataRetentionTime)
		if _, err := c.minioClient.PutObject(
			c.config.App.Storage.Bucket,
			object,
			r,
			size,
			minio.PutObjectOptions{
				ContentType: ": image/jpeg",
				//RetainUntilDate:&t,
			},
		); err != nil {
			log.Error(err)
			continue
		}

		grpc_server.ServerStats.Add(grpc_server.StatsCount, 1)
		grpc_server.ServerStats.Add(grpc_server.StatsSize, size)
		grpc_server.ServerStats.Add(grpc_server.StatsWorkingTime, time.Since(started).Milliseconds())
	}
	return nil
}
