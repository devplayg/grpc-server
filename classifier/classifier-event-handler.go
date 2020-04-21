package classifier

import (
	"bytes"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type EventWrapper struct {
	event *proto.Event
	Uuid  uuid.UUID
}

func (c *Classifier) handleEvent() error {
	go func() {
		batch := make([]*EventWrapper, 0, c.batchSize)
		timer := time.NewTimer(c.batchTimeout)
		timer.Stop()

		save := func() {
			if err := c.save(batch); err != nil {
				log.Error(err)
			}
			batch = make([]*EventWrapper, 0, c.batchSize)
		}

		for {
			select {
			case event := <-c.eventCh:
				eventWrapper := &EventWrapper{event: event}
				eventWrapper.Uuid = uuid.New()
				batch = append(batch, eventWrapper)
				if len(batch) == 1 {
					timer.Reset(c.batchTimeout)
				}
				if len(batch) == c.batchSize {
					timer.Stop()
					save()
				}
			case <-c.Ctx.Done():
				log.Debug("storage channel is done")
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

func (c *Classifier) eventToTsv(events []*EventWrapper) string {
	var text string
	for _, r := range events {
		deviceId, _ := c.deviceCodeMap[r.event.Header.DeviceCode]
		text += fmt.Sprintf("%s\t%d\t%d\t%s\t%d\n",
			time.Unix(r.event.Header.Date.Seconds, 0).Format(grpc_server.DefaultDateFormat),
			deviceId,
			r.event.Header.EventType,
			r.Uuid.String(),
			0,
		)
	}
	return strings.TrimSpace(text)
}

func writeTextIntoTempFile(dir, text string) (string, error) {
	tmpFile, err := ioutil.TempFile(dir, "db-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file for saving data; %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(text); err != nil {
		return "", fmt.Errorf("failed to write data into temp file; %w", err)
	}
	return filepath.ToSlash(tmpFile.Name()), nil
}

func (c *Classifier) save(events []*EventWrapper) error {
	if err := c.saveHeader(events); err != nil {
		return fmt.Errorf("failed to insert; %w", err)
	}

	if err := c.saveBody(events); err != nil {
		return fmt.Errorf("failed to insert; %w", err)
	}

	return nil
}

func (c *Classifier) saveHeader(events []*EventWrapper) error {
	started := time.Now()
	text := c.eventToTsv(events)
	path, err := writeTextIntoTempFile(grpc_server.TempDir, text)
	if err != nil {
		return err
	}
	db := grpc_server.BulkInsert(c.db, "log", "date, device_id, event_type, UUID, flag", path)
	if db.Error != nil {
		return err
	}
	os.Remove(path)
	dur := time.Since(started)
	log.Debugf("inserted %d row(s)", db.RowsAffected)
	log.WithFields(logrus.Fields{
		"time":         dur.Seconds(),
		"rowsAffected": db.RowsAffected,
	}).Debugf("inserted")

	stats.Add("inserted", db.RowsAffected)
	stats.Add("inserted-time", dur.Milliseconds())
	return nil
}

func (c *Classifier) saveBody(events []*EventWrapper) error {
	started := time.Now()
	var total int64

	r := bytes.NewReader(nil)
	for _, e := range events {
		for i, f := range e.event.Body.Files {
			size := int64(len(f.Data))
			r.Reset(f.Data)
			started := time.Now()
			_, err := c.minioClient.PutObject(
				c.config.App.Storage.Bucket,
				fmt.Sprintf("%s_%d.jpg", e.Uuid.String(), i),
				r,
				size,
				minio.PutObjectOptions{ContentType: ": image/jpeg"},
			)
			if err != nil {
				log.Error(err)
				continue
			}

			stats.Add("uploaded", 1)
			stats.Add("uploaded-size", size)
			stats.Add("uploaded-time", time.Since(started).Milliseconds())
			total += size
		}
	}

	log.WithFields(logrus.Fields{
		"time":    time.Since(started).Seconds(),
		"count":   len(events),
		"size(B)": total,
		"Bps":     float32(total) / float32(len(events)),
	}).Debugf("saved")

	return nil
}
