package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

func (c *Classifier) handleEvent() error {
	go func() {
		batch := make([]*proto.Event, 0, c.batchSize)
		timer := time.NewTimer(c.batchTimeout)
		timer.Stop()

		save := func() {
			var text string
			for _, r := range batch {
				deviceId, _ := c.deviceCodeMap[r.Header.DeviceCode]
				text += fmt.Sprintf("%s\t%d\t%d\t%s\t%d\n",
					time.Unix(r.Header.Date.Seconds, 0).Format(grpc_server.DefaultDateFormat),
					deviceId,
					r.Header.EventType,
					uuid.New().String(),
					0,
				)
			}
			text = strings.TrimSpace(text)
			tmpFile, err := ioutil.TempFile(grpc_server.TempDir, "db-")
			if err != nil {
				log.Error(fmt.Errorf("failed to create temporary file for saving data; %w", err))
				return
			}
			if _, err := tmpFile.WriteString(text); err != nil {
				log.Error(fmt.Errorf("failed to write data into temp file; %w", err))
				return
			}
			tmpFile.Close()

			db := c.insertFactoryEvent(tmpFile.Name())
			if db.Error != nil {
				log.Error(fmt.Errorf("failed to insert; %w", err))
				return
			}
			log.Debugf("inserted %d row(s)", db.RowsAffected)

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

func (c *Classifier) insertFactoryEvent(path string) *gorm.DB {
	p := filepath.ToSlash(path)
	query := `
		LOAD DATA LOCAL INFILE %q
		INTO TABLE log
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
		(DATE, device_id, event_type, UUID, flag)`
	query = fmt.Sprintf(query, p)
	return c.db.Exec(query)

}
