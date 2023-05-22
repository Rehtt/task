
make-rpc:
	protoc --proto_path=./rpc --go_out=paths=source_relative:./rpc --go-grpc_out=paths=source_relative:./rpc ./rpc/*.proto

make-client:
	go build -o bin/client.exe ./cmd/client