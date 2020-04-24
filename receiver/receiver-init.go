package receiver

import (
	"context"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"github.com/devplayg/grpc-server/classifier"
	"net/http"
	"os"
	"time"
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

func (r *Receiver) startMonitor() error {
	srv := http.Server{
		Addr: r.monitorAddr,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			ctx = context.WithValue(ctx, "err", err)
			cancel()
			return
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Value("err").(error)
	case <-r.Ctx.Done():
		log.Debug(fmt.Errorf("monitoring service received stop signal from server"))
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutDown); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}
