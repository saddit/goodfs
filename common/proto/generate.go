package proto

//go:generate protoc -I=. --go_out=. --go-grpc_out=. raft_cmd.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. message.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. hashslot.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. metadata.proto
