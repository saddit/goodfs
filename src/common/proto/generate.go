package proto

//sudo apt install -y protobuf-compiler
//go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
//go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
//go:generate protoc -I=. --go_out=. --go-grpc_out=. raft_cmd.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. message.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. hashslot.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. metadata.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. object_migration.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. config_service.proto
