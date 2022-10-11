package entity

import "net"

//go:generate msgp -tests=false

type PeerInfo struct {
	ServerID string `msg:"server_id"`
	Location string `msg:"location"`
	HttpPort string `msg:"http_port"`
	RpcPort  string `msg:"gprc_port"`
}

func (p *PeerInfo) RpcAddress() string {
	return net.JoinHostPort(p.Location, p.RpcPort)
}

func (p *PeerInfo) HttpAddress() string {
	return net.JoinHostPort(p.Location, p.HttpPort)
}