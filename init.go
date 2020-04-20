package grpc_server

func init() {
	if err := EnsureDir(DataDir); err != nil {
		panic(err)
	}

	if err := EnsureDir(TempDir); err != nil {
		panic(err)
	}
}
