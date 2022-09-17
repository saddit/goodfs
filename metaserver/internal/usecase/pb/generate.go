package pb

//go:generate protoc -I=proto --go_out=. --go-grpc_out=. proto/raft_cmd.proto
//go:generate protoc -I=proto --go_out=. --go-grpc_out=. proto/message.proto
//go:generate protoc -I=proto --go_out=. --go-grpc_out=. proto/hashslot.proto
