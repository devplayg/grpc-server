package grpc_server

import (
	"expvar"
	"github.com/devplayg/grpc-server/proto"
	"runtime"
	"time"
)

const (
	DataDir = "./data"
	TempDir = "./tmp"

	StatsInitialProcessing = "initialProcessing"
	StatsLastProcessing    = "lastProcessing"
	StatsCount             = "count"
	StatsSize              = "size"
	StatsWorker            = "worker"
	StatsWorkingTime       = "workingTime"

	// DefaultDateFormat = "2006-01-02 15:04:05"
)

var (
	ServerStats *expvar.Map
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ServerStats = expvar.NewMap("server")

	if err := EnsureDir(DataDir); err != nil {
		panic(err)
	}

	if err := EnsureDir(TempDir); err != nil {
		panic(err)
	}
}

func ResetServerStats(extras ...string) {
	ServerStats.Set(StatsInitialProcessing, new(expvar.Int))
	ServerStats.Get(StatsInitialProcessing).(*expvar.Int).Set(time.Now().UnixNano())
	ServerStats.Set(StatsLastProcessing, new(expvar.Int))
	ServerStats.Set(StatsCount, new(expvar.Int))
	ServerStats.Set(StatsSize, new(expvar.Int))
	ServerStats.Set(StatsWorker, new(expvar.Int))
	ServerStats.Set(StatsWorkingTime, new(expvar.Int))

	for _, e := range extras {
		ServerStats.Set(e, new(expvar.Int))
	}
}

func GetServerStats() *proto.ServerStats {
	return &proto.ServerStats{
		StartTimeUnixNano:     ServerStats.Get(StatsInitialProcessing).(*expvar.Int).Value(),
		EndTimeUnixNano:       ServerStats.Get(StatsLastProcessing).(*expvar.Int).Value(),
		Count:                 ServerStats.Get(StatsCount).(*expvar.Int).Value(),
		Size:                  ServerStats.Get(StatsSize).(*expvar.Int).Value(),
		Worker:                int32(ServerStats.Get(StatsWorker).(*expvar.Int).Value()),
		TotalWorkingTimeMilli: ServerStats.Get(StatsWorkingTime).(*expvar.Int).Value(),
		Meta:                  "",
	}
}
