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
		Receiver struct {
			Address    string
			Classifier struct {
				Address string
			}
		}
		Classifier struct {
			Address string
		}
		Notifier struct {
			Address string
		}
	}

	Spring struct {
		DataSource struct {
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

		//config, err := readSpringBootConfig(p)
		//if err != nil {
		//	return nil, err
		//}

		//if config.App.VmsServer == nil {
		//	return nil, errors.New("VMS-configuration not found")
		//}

		//if config.App.VmsServer == nil {
		//	return nil, errors.New("VAS-configuration not found")
		//}

		//config.App.VmsServer.Tune()
		//for _, v := range config.App.VasServer {
		//	v.Tune()
		//}

		//return config, nil
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
