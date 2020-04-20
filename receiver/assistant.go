package receiver

import (
	"fmt"
	"github.com/devplayg/golibs/converter"
	"github.com/devplayg/grpc-server/proto"
	"io/ioutil"
	"os"
	"time"
)

type assistant struct {
	storageCh       chan *proto.Event
	maxQueueSize    int
	processInterval time.Duration
	storage         string
}

func newAssistant(batchSize int, batchTimeout time.Duration, storage string) *assistant {
	return &assistant{
		storageCh:       make(chan *proto.Event, batchSize),
		maxQueueSize:    batchSize,
		processInterval: batchTimeout,
		storage:         storage,
	}
}

func (a *assistant) Start() error {
	fi, err := os.Stat(a.storage)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(a.storage, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory; %w", err)
		}
	}
	if !fi.IsDir() {
		return fmt.Errorf("storage '%s' is not directory", fi.Name())
	}

	go func() {
		batch := make([]*proto.Event, 0, a.maxQueueSize)
		timer := time.NewTimer(a.processInterval)
		timer.Stop()

		save := func() {
			encoded, err := converter.EncodeToBytes(batch)
			if err != nil {
				log.Error(fmt.Errorf("failed to encode bytes; %w", err))
				return
			}
			f, err := ioutil.TempFile(a.storage, "receiver-data-")
			if err != nil {
				log.Error(err)
				return
			}
			defer f.Close()
			f.Write(encoded)
			log.Debugf("saved %d", len(batch))
			batch = make([]*proto.Event, 0, a.maxQueueSize)
		}

		for {
			select {
			case event := <-a.storageCh:
				batch = append(batch, event)
				if len(batch) == 1 {
					timer.Reset(a.processInterval)
				}
				if len(batch) == a.maxQueueSize {
					timer.Stop()
					save()
				}
			case <-timer.C:
				// stats.Add("batchTimeout", 1)
				save()
			}
		}
	}()

	return nil
}
