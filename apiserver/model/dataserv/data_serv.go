package dataserv

import (
	"sync/atomic"
	"time"
)

type ServState int8

const (
	Healthy ServState = iota
	Suspend
)

type DataServ struct {
	Ip       string
	LastBeat time.Time
	state    atomic.Value
}

func New(ip string) *DataServ {
	s := atomic.Value{}
	s.Store(Healthy)
	return &DataServ{
		Ip:       ip,
		LastBeat: time.Now(),
		state:    s,
	}
}

func (d *DataServ) IsAvailable() bool {
	return d.GetState() == Healthy
}

func (d *DataServ) GetState() ServState {
	return d.state.Load().(ServState)
}

func (d *DataServ) SetState(state ServState) {
	d.state.Store(state)
}