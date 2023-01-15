package entity

//go:generate msgp -tests=false

type PeerInfo struct {
	Location string `msg:"location"`
	HttpPort string `msg:"http_port"`
	GrpcPort string `msg:"grpc_port"`
	GroupID  string `msg:"group_id"`
}
