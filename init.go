package grpc_server

import "net/http"

func init() {
	if err := EnsureDir(DataDir); err != nil {
		panic(err)
	}

	if err := EnsureDir(TempDir); err != nil {
		panic(err)
	}

	go http.ListenAndServe(":8123", nil)
}
