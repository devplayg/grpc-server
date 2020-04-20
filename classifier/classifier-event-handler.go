package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (c *Classifier) eventToTsv(events []*proto.Event) string {
	var text string
	for _, r := range events {
		deviceId, _ := c.deviceCodeMap[r.Header.DeviceCode]
		text += fmt.Sprintf("%s\t%d\t%d\t%s\t%d\n",
			time.Unix(r.Header.Date.Seconds, 0).Format(grpc_server.DefaultDateFormat),
			deviceId,
			r.Header.EventType,
			uuid.New().String(),
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

func (c *Classifier) saveIntoDatabase(events []*proto.Event) error {
	text := c.eventToTsv(events)
	path, err := writeTextIntoTempFile(grpc_server.TempDir, text)
	if err != nil {
		return err
	}
	db := grpc_server.BulkInsert(c.db, "log", "date, device_id, event_type, UUID, flag", path)
	if db.Error != nil {
		return fmt.Errorf("failed to insert; %w", err)
	}
	log.Debugf("inserted %d row(s)", db.RowsAffected)
	os.Remove(path)
	return nil
}

func (c *Classifier) handleEvent() error {
	go func() {
		batch := make([]*proto.Event, 0, c.batchSize)
		timer := time.NewTimer(c.batchTimeout)
		timer.Stop()

		save := func() {
			if err := c.saveIntoDatabase(batch); err != nil {
				log.Error(err)
			}
			batch = make([]*proto.Event, 0, c.batchSize)
		}

		for {
			select {
			case event := <-c.eventCh:
				batch = append(batch, event)
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
