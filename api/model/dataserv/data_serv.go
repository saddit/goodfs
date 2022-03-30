package dataserv

import (
	"time"
)

type ServState int8

const (
	Dead ServState = iota
	Healthy
	Suspend
)

type DataServ struct {
	Ip       string
	LastBeat time.Time
	State    ServState
}

func New(ip string) *DataServ {
	return &DataServ{
		Ip:       ip,
		LastBeat: time.Now(),
		State:    Healthy,
	}
}

func (d *DataServ) IsAvailable() bool {
	return d.State == Healthy
}
