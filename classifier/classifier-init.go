package classifier

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
)

func (c *Classifier) init() error {
	log = c.Log

	config, err := grpc_server.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration")
	}
	if len(config.App.Classifier.Address) < 1 {
		config.App.Classifier.Address = "127.0.0.1:8802"
	}

	c.config = config
	return nil
}
