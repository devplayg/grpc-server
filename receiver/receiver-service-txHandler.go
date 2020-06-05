package receiver

import (
	"fmt"
	"github.com/devplayg/goutils"
	"github.com/devplayg/grpc-server/proto"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"sync"
	"time"
)

func (r *Receiver) startTxHandler(wg *sync.WaitGroup, serviceName string) {
	go func() {
		log.WithFields(logrus.Fields{
			"batchSize":        r.batchSize,
			"batchTimeout(ms)": r.batchTimeout.Milliseconds(),
		}).Debugf("%s has been started", serviceName)

		defer func() {
			log.Debugf("%s has been stopped", serviceName)
			wg.Done()
		}()
		if err := r._startTxHandler(); err != nil {
			log.Error(fmt.Errorf("failed to run %s: %w", serviceName, err))
			r.Cancel()
			return
		}
	}()
}

func (r *Receiver) _startTxHandler() error {
	ch := make(chan struct{})
	go func() {
		defer close(ch)

		batch := make([]*proto.Event, 0, r.batchSize)
		timer := time.NewTimer(r.batchTimeout)
		timer.Stop()

		save := func() {
			defer func() {
				batch = make([]*proto.Event, 0, r.batchSize)
			}()
			if err := saveEvents(batch, r.storage); err != nil {
				log.Error(err)
				return
			}
			log.Debugf("saved %d", len(batch))
		}

		for {
			select {
			case event := <-r.storageCh:
				batch = append(batch, event)
				if len(batch) == 1 {
					timer.Reset(r.batchTimeout)
				}
				if len(batch) == r.batchSize {
					timer.Stop()
					save()
				}
			case <-r.Ctx.Done():
				log.Debug("tx-failed-handler received stop signal")
				if len(batch) > 0 {
					save()
				}
				return
			case <-timer.C:
				save()
			}
		}
	}()
	<-ch
	return nil
}

func saveEvents(events []*proto.Event, dir string) error {
	encoded, err := goutils.GobEncode(events)
	if err != nil {
		return fmt.Errorf("failed to encode bytes; %w", err)
	}
	f, err := ioutil.TempFile(dir, "receiver-data-")
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(encoded); err != nil {
		return err
	}
	return nil

}
