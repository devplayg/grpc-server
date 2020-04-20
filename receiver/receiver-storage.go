package receiver

import (
	"fmt"
	"github.com/devplayg/golibs/converter"
	"github.com/devplayg/grpc-server/proto"
	"io/ioutil"
	"os"
	"time"
)

func (r *Receiver) runStorageCh() error {
	fi, err := os.Stat(r.storage)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(r.storage, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory; %w", err)
		}
	}
	if !fi.IsDir() {
		return fmt.Errorf("storage '%s' is not directory", fi.Name())
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
			case <-timer.C:
				save()
			}
		}
	}()

	return nil
}
