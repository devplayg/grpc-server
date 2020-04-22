package classifier

import (
	"bytes"
	"expvar"
	"fmt"
	"github.com/devplayg/grpc-server/entity"
	"github.com/devplayg/grpc-server/proto"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
	"sync"
	"time"
)

type EventWrapper struct {
	event    *proto.Event
	uuid     uuid.UUID
	flag     int
	deviceId int64
}

func (e EventWrapper) entity() *entity.Log {
	eventType, _ := proto.EventHeader_EventType_value[e.event.Header.EventType.String()]
	return &entity.Log{
		Date:        time.Unix(e.event.Header.Date.Seconds, 0),
		DeviceId:    e.deviceId,
		EventType:   int(eventType),
		Uuid:        e.uuid.String(),
		Flag:        e.flag,
		AttachCount: len(e.event.Body.Files),
	}
}

func (c *Classifier) save(event *proto.Event) error {
	c.once.Do(func() {
		stats.Get("start").(*expvar.Int).Set(time.Now().UnixNano())
	})
	deviceId, _ := c.deviceCodeMap[event.Header.DeviceCode]
	eventWrapper := &EventWrapper{
		event:    event,
		uuid:     uuid.New(),
		deviceId: deviceId,
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		//if err := c.db.Create(&event).Error; err != nil {
		//	log.Error(fmt.Errorf("failed to insert; %w", err))
		//	return
		//}
		if err := c.saveHeader(eventWrapper); err != nil {
			log.Error(fmt.Errorf("failed to insert; %w", err))
			return
		}
	}()

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
	wg.Wait()

	stats.Get("end").(*expvar.Int).Set(time.Now().UnixNano())
	return nil
}

//
//func (c *Classifier) protoEventToEntity(e *EventWrapper) *entity.Log {
//	deviceId, _ := c.deviceCodeMap[e.event.Header.DeviceCode]
//	log.Debug(e.event.Header.EventType.String())
//
//	return &entity.Log{
//		Date:        time.Unix(e.event.Header.Date.Seconds, 0),
//		DeviceId:    deviceId,
//		EventType:   3,
//		Uuid:        e.uuid.String(),
//		Flag:        e.flag,
//		AttachCount: len(e.event.Body.Files),
//	}
//}

func (c *Classifier) saveHeader(event *EventWrapper) error {
	started := time.Now()
	db := c.db.Create(event.entity())
	if db.Error != nil {
		return db.Error
	}
	dur := time.Since(started)
	stats.Add("inserted", db.RowsAffected)
	stats.Add("inserted-time", dur.Milliseconds())

	//c.db.Create(&event)
	//text := c.eventToTsv(events)
	//path, err := writeTextIntoTempFile(grpc_server.TempDir, text)
	//if err != nil {
	//	return err
	//}
	//db := grpc_server.BulkInsert(c.db, "log", "date, device_id, event_type, UUID, flag", path)
	//if db.Error != nil {
	//	return err
	//}
	//os.Remove(path)
	//log.WithFields(logrus.Fields{
	//	"insert-time": dur.Seconds(),
	//	"count":       db.RowsAffected,
	//}).Debugf("inserted")
	//
	return nil
}

func (c *Classifier) saveBody(e *EventWrapper) error {
	//started := time.Now()
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

		stats.Add("uploaded", 1)
		stats.Add("uploaded-size", size)
		stats.Add("uploaded-time", time.Since(started).Milliseconds())
		total += size
	}
	return nil
}

//
//func (c *Classifier) handleEvent() error {
//	//ch := make(chan bool, c.workerCount)
//	//go func() {
//		//for {
//		//	select {
//		//	case event := <-c.eventCh:
//		//		//c.insert(event.Header)
//		//	case <-c.Ctx.Done():
//		//		log.Debug("storage channel is done")
//		//
//		//		return
//		//
//		//	}
//		//}
//	//}()
//
//	return nil
//}
//
//func (c *Classifier) eventToTsv(events []*EventWrapper) string {
//	var text string
//	for _, r := range events {
//		deviceId, _ := c.deviceCodeMap[r.event.Header.DeviceCode]
//		text += fmt.Sprintf("%s\t%d\t%d\t%s\t%d\n",
//			time.Unix(r.event.Header.Date.Seconds, 0).Format(grpc_server.DefaultDateFormat),
//			deviceId,
//			r.event.Header.EventType,
//			r.Uuid.String(),
//			0,
//		)
//	}
//	return strings.TrimSpace(text)
//}

//
//func writeTextIntoTempFile(dir, text string) (string, error) {
//	tmpFile, err := ioutil.TempFile(dir, "db-")
//	if err != nil {
//		return "", fmt.Errorf("failed to create temporary file for saving data; %w", err)
//	}
//	defer tmpFile.Close()
//
//	if _, err := tmpFile.WriteString(text); err != nil {
//		return "", fmt.Errorf("failed to write data into temp file; %w", err)
//	}
//	return filepath.ToSlash(tmpFile.Name()), nil
//}
