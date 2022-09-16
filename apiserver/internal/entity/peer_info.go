package entity

//go:generate msgp

type PeerInfo struct {
	Location string `msg:"location"`
	HttpPort string `msg:"http_port"`
	GrpcPort string `msg:"gprc_port"`
	GroupID  string `msg:"group_id"`
}
