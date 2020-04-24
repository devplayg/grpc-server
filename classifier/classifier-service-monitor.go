package classifier

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

func (c *Classifier) startMonitoringService(wg *sync.WaitGroup, serviceName string) {
	go func() {
		log.WithFields(logrus.Fields{
			"address": c.monitorAddr,
		}).Debugf("%s has been started", serviceName)
		defer func() {
			log.Debugf("%s has been stopped", serviceName)
			wg.Done()
		}()

		if err := c._startMonitor(); err != nil {
			log.Error(fmt.Errorf("failed to start %s: %w", serviceName, err))
			c.Cancel()
			return
		}
	}()
}

func (c *Classifier) _startMonitor() error {
	srv := http.Server{
		Addr: c.monitorAddr,
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

	case <-c.Ctx.Done(): // from receiver context
		log.Debug(fmt.Errorf("monitoring service received stop signal"))
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutDown); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}
