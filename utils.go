package grpc_server

import (
	"fmt"
	"os"
)

func EnsureDir(dir string) error {
	fi, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory; %w", err)
		}
	}
	if !fi.IsDir() {
		return fmt.Errorf("storage '%s' is not directory", fi.Name())
	}

	return nil
}
