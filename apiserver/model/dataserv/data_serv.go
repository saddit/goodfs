package dataserv

import (
	"log"
	"sync/atomic"
	"time"
)

type ServState int8

const (
	Healthy ServState = iota
	Suspend
	Death
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
	if r := d.state.Load(); r != nil {
		return r.(ServState)
	}
	log.Println("Error: atomic.state return nil..")
	return Death
}

func (d *DataServ) SetState(state ServState) {
	d.state.Store(state)
}
