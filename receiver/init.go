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
	if len(config.App.Receiver.BindAddress) < 1 {
		config.App.Receiver.BindAddress = "127.0.0.1:8801"
	}

	r.config = config
	return nil
}
