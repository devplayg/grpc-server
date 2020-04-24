package receiver

import (
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/classifier"
	"os"
	"strings"
)

func (r *Receiver) init() error {
	log = r.Log
	grpc_server.ResetServerStats()

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
		config.App.Receiver.Address = DefaultAddress
	}

	if len(config.App.Receiver.StorageDir) < 1 {
		config.App.Receiver.StorageDir = DefaultStorageDir
	}

	if len(config.App.Receiver.Classifier.Address) < 1 {
		config.App.Receiver.Classifier.Address = classifier.DefaultAddress
	}
	arr := strings.Split(config.App.Receiver.Classifier.Address, ":")
	if len(arr) == 2 && len(arr[0]) < 1 {
		config.App.Receiver.Classifier.Address = "127.0.0.1:" + arr[1]
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
