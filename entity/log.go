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

//func (f LogEvent) ToTsv(loc *time.Location) string {
//	return fmt.Sprintf("%s\t%d\t%d\t%d\t%d\t%s\t%v\t%s\n",
//		f.Date.In(loc).Format("2006-01-02 15:04:05"),
//		f.FactoryId,
//		f.CameraId,
//		f.EventType,
//		f.EventType,
//		f.Path,
//		f.Archived,
//		f.TargetId,
//	)
//}

func (Log) TableName() string {
	return "log"
}
