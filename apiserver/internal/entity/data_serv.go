package entity

import (
	"log"
	"sync/atomic"
	"time"
)

type ServState int8

const (
	ServStateHealthy ServState = iota
	ServStateSuspend
	ServStateDeath
)

type DataServ struct {
	Ip       string
	LastBeat time.Time
	state    atomic.Value
}

func NewDataServ(ip string) *DataServ {
	s := atomic.Value{}
	s.Store(ServStateHealthy)
	return &DataServ{
		Ip:       ip,
		LastBeat: time.Now(),
		state:    s,
	}
}

func (d *DataServ) IsAvailable() bool {
	return d.GetState() == ServStateHealthy
}

func (d *DataServ) GetState() ServState {
	if r := d.state.Load(); r != nil {
		return r.(ServState)
	}
	log.Println("Error: atomic.state return nil..")
	return ServStateDeath
}

func (d *DataServ) SetState(state ServState) {
	d.state.Store(state)
}
