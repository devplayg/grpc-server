package receiver

import (
	"expvar"
	"fmt"
	grpc_server "github.com/devplayg/grpc-server"
	"net/http"
	"os"
	"time"
)

func (r *Receiver) init() error {
	log = r.Log
	stats = expvar.NewMap(r.Engine.Config.Name)

	if err := r.loadConfig(); err != nil {
		return err
	}

	if err := r.initCredentials(); err != nil {
		return err
	}

	if err := r.initMonitor(); err != nil {
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

func (r *Receiver) initMonitor() error {
	resetStats()

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		m := map[string]interface{}{
			"duration": (stats.Get("end").(*expvar.Int).Value() - stats.Get("start").(*expvar.Int).Value()) / int64(time.Millisecond),
			"relayed":  stats.Get("relayed").(*expvar.Int).Value(),
			"size":     stats.Get("size").(*expvar.Int).Value(),
		}
		s := fmt.Sprintf("%d\t%d\t%d", m["relayed"], m["duration"], m["size"])
		w.Write([]byte(s))
	})

	go http.ListenAndServe(":8123", nil)

	return nil
}

func resetStats() {
	log.Debug("reset stats")

	stats.Set("start", new(expvar.Int))
	stats.Get("start").(*expvar.Int).Set(time.Now().UnixNano())
	stats.Set("end", new(expvar.Int))
	stats.Set("relayed", new(expvar.Int))
	stats.Set("size", new(expvar.Int))
}
