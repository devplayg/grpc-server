package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
)

func (r *Receiver) init() error {
	config, err := grpc_server.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration")
	}
	if len(config.App.Receiver.Address) < 1 {
		config.App.Receiver.Address = "127.0.0.1:8801"
	}
	if len(config.App.Receiver.Classifier.Address) < 1 {
		config.App.Receiver.Classifier.Address = "127.0.0.1:8802"
	}

	r.config = config
	return nil
}
