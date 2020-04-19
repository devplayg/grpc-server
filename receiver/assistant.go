package receiver

import (
	"fmt"
	"github.com/devplayg/grpc-server/proto"
	"io/ioutil"
	"time"
)

type assistant struct {
	ch              chan *proto.Event
	maxQueueSize    int
	processInterval time.Duration
}

func newAssistant(maxQueueSize int, processInterval time.Duration) *assistant {
	return &assistant{
		ch:              make(chan *proto.Event, maxQueueSize),
		maxQueueSize:    maxQueueSize,
		processInterval: processInterval,
	}
}

func (a *assistant) Start() error {
	go func() {
		batch := make([]*proto.Event, 0, a.maxQueueSize)
		timer := time.NewTimer(a.processInterval)
		timer.Stop()

		save := func() {
			f, err := ioutil.TempFile(".", "receiver-data-")
			if err != nil {
				log.Error(err)
				return
			}
			for _, b := range batch {
				f.WriteString(fmt.Sprintf("%v\t%d\n", b.Header.Date, b.Header.RiskLevel))
			}
			f.Close()
			log.Debugf("saved - %d", len(batch))
			batch = make([]*proto.Event, 0, a.maxQueueSize)
		}

		for {
			select {
			case event := <-a.ch:
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
