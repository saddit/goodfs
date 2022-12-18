package entity

//go:generate msgp -tests=false

type PeerInfo struct {
	Location string `msg:"location" json:"location"`
	HttpPort string `msg:"http_port" json:"httpPort"`
	GrpcPort string `msg:"gprc_port" json:"grpcPort"`
	GroupID  string `msg:"group_id" json:"groupId"`
	ServerID string `yaml:"server_id" json:"serverId"`
}
