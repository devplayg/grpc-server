package entity

import "time"

type Log struct {
	LogId       int64 `gorm:"primary_key"`
	Date        time.Time
	DeviceId    int64
	EventType   int
	Uuid        string
	Flag        int
	AttachCount int
}

func (Log) TableName() string {
	return "log"
}
