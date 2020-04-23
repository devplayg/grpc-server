package classifier

import (
	"github.com/devplayg/grpc-server/entity"
	"github.com/devplayg/grpc-server/proto"
	"github.com/google/uuid"
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
