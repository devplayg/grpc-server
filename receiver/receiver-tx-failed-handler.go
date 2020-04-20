package receiver

import (
	"fmt"
	"github.com/devplayg/golibs/converter"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/proto"
	"io/ioutil"
	"time"
)

func (r *Receiver) handleTxFailedEvent() error {
	if err := grpc_server.EnsureDir(r.storage); err != nil {
		return err
	}

	go func() {
		batch := make([]*proto.Event, 0, r.batchSize)
		timer := time.NewTimer(r.batchTimeout)
		timer.Stop()

		save := func() {
			encoded, err := converter.EncodeToBytes(batch)
			if err != nil {
				log.Error(fmt.Errorf("failed to encode bytes; %w", err))
				return
			}
			f, err := ioutil.TempFile(r.storage, "receiver-data-")
			if err != nil {
				log.Error(err)
				return
			}
			defer f.Close()
			f.Write(encoded)
			log.Debugf("saved %d", len(batch))
			batch = make([]*proto.Event, 0, r.batchSize)
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
