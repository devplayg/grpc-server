## Generate code

    protoc -I . --go_out=plugins=grpc:. event.proto