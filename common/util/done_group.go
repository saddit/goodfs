package util

import (
	"sync"

	"go.uber.org/atomic"
)

type NonErrDoneGroup interface {
	Add(int)
	Done()
	Wait()
	WaitDone() <-chan struct{}
	Todo()
}

type DoneGroup struct {
	sync.WaitGroup
	ec chan error
	closed *atomic.Bool
}

// NewNonErrDoneGroup equals to WaitGroup. Only Todo() and WaitDone() func can be used!
func NewNonErrDoneGroup() NonErrDoneGroup {
	return &DoneGroup{sync.WaitGroup{}, nil, atomic.NewBool(false)}
}

func NewDoneGroup() DoneGroup {
	return DoneGroup{sync.WaitGroup{}, make(chan error, 1), atomic.NewBool(false)}
}

//Todo equals to wg.Add(1)
func (d *DoneGroup) Todo() {
	d.Add(1)
}

//Error deliver an error non blocking
func (d *DoneGroup) Error(e error) {
	if d.closed.Load() {
		return
	}
	d.ec <- e
}

func (d *DoneGroup) WaitError() <-chan error {
	return d.ec
}

func (d *DoneGroup) WaitDone() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		d.Wait()
		ch <- struct{}{}
	}()
	return ch
}

//Close close the error chan
func (d *DoneGroup) Close() {
	close(d.ec)
	d.closed.CAS(false, true)
}

//WaitUntilError use select to WaitDone() and WaitError() if has error return it else return nil
func (d *DoneGroup) WaitUntilError() error {
	for {
		select {
		case <-d.WaitDone():
			return nil
		case e := <-d.WaitError():
			if e != nil {
				return e
			}
		}
	}
}
