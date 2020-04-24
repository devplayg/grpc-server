package receiver

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

func (r *Receiver) startMonitoringService(wg *sync.WaitGroup) {
	serviceName := "monitoring service"

	go func() {
		log.WithFields(logrus.Fields{
			"address": r.monitorAddr,
		}).Debugf("%s has been started", serviceName)
		defer func() {
			log.Debugf("%s has been stopped", serviceName)
			wg.Done()
		}()

		if err := r._startMonitor(); err != nil {
			log.Error(fmt.Errorf("failed to start %s: %w", serviceName, err))
			r.Cancel()
			return
		}
	}()
}

func (r *Receiver) _startMonitor() error {
	srv := http.Server{
		Addr: r.monitorAddr,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			ctx = context.WithValue(ctx, "err", err)
			cancel()
			return
		}
	}()

	select {
	case <-ctx.Done(): // from local context
		return ctx.Value("err").(error)

	case <-r.Ctx.Done(): // from receiver context
		log.Debug(fmt.Errorf("monitoring service received stop signal from server"))
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutDown); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}
