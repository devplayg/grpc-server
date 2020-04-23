package grpc_server

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
)

var configPaths = []string{
	"config.yaml",
	"conf/config.yaml",
	"../client/config/application-dev.yaml",
}

type Config struct {
	App struct {

		// Receiver
		Receiver struct {
			Insecure   bool
			StorageDir string `json:"storage-dir"`
			Address    string
			Classifier struct {
				Address string
			}
		}

		// Classifier
		Classifier struct {
			Address  string
			Insecure bool
			Notifier struct {
				Address string
			}
		}

		// Notifier
		Notifier struct {
			Address  string
			Insecure bool
		}

		Storage struct {
			Address   string
			Bucket    string
			AccessKey string `json:"access-key"`
			SecretKey string `json:"secret-key"`
		}
	}

	Spring struct {
		Datasource struct {
			Url      string
			Username string
			Password string
		}
	}
}

func LoadConfig() (*Config, error) {
	for _, path := range configPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		config, err := readConfigFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read configuration file: %w", err)
		}
		return config, nil
	}
	return nil, nil
}

func readConfigFile(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
