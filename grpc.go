package grpc_server

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"log"
	"reflect"
	"time"
)

type ConnStatsHandler struct {
	From string
	To   string
	Log  *logrus.Logger
}

func (c *ConnStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	ctx = context.WithValue(ctx, "remoteaddr", info.RemoteAddr.String())
	//fmt.Printf("TagConn\tinfo.RemoteAddr=%v\n", info.RemoteAddr)
	return ctx
}
func (c *ConnStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	//fmt.Printf("[%s] TagRPC\tinfo.FullMethodName=%s\tinfo.FailFast=%v\n", getCtxValue(ctx), info.FullMethodName, info.FailFast)
	return ctx
}
func (c *ConnStatsHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	//fmt.Printf("[%s] HandleRPC\trpcStats.IsClient()=%v\n", getCtxValue(ctx), rpcStats.IsClient())
}
func (c *ConnStatsHandler) HandleConn(ctx context.Context, connStats stats.ConnStats) {
	if reflect.TypeOf(connStats).String() == "*stats.ConnBegin" {
		//fmt.Printf("[%s-%s] connected\n", s.Name, getCtxValue(ctx))
		c.Log.WithFields(logrus.Fields{
			"from": c.From,
			"to":   c.To,
			"addr": getCtxValue(ctx),
		}).Info("connected")
		return
	}
	//fmt.Printf("[%s-%s] disconnected\n", s.Name, getCtxValue(ctx))
	c.Log.WithFields(logrus.Fields{
		"from": c.From,
		"to":   c.To,
		"addr": getCtxValue(ctx),
	}).Info("disconnected")
}

func getCtxValue(ctx context.Context) string {
	if v := ctx.Value("remoteaddr"); v != nil {
		u, ok := v.(string)
		if ok {
			return u
		}
	}
	return ""
}

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	//
	//skip auth when ListReleases requested
	//if info.FullMethod != "/proto.GoReleaseService/ListReleases" {
	//	if err := authorize(ctx); err != nil {
	//		return nil, err
	//	}
	//}
	h, err := handler(ctx, req)

	//logging
	log.Printf("request - Method:%s\tDuration:%s\tError:%v\n",
		info.FullMethod,
		time.Since(start),
		err)

	return h, err
}
