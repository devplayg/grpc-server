package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"os"
)

func (r *Receiver) init() error {
	if err := r.loadConfig(); err != nil {
		return err
	}

	if err := r.initCredentials(); err != nil {
		return err
	}
	return nil
}

func (r *Receiver) loadConfig() error {
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

func (r *Receiver) initCredentials() error {
	if r.Engine.Config.Insecure {
		return nil
	}

	if _, err := os.Stat(r.Engine.Path(r.Engine.Config.CertFile)); os.IsNotExist(err) {
		return fmt.Errorf("certificate file not found; %w", err)
	}
	if _, err := os.Stat(r.Engine.Path(r.Engine.Config.KeyFile)); os.IsNotExist(err) {
		return fmt.Errorf("certificate key file not found; %w", err)
	}

	return nil
}
