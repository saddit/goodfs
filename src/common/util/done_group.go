package util

import (
	"common/graceful"
	"errors"
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
	ec     chan error
	closed *atomic.Bool
}

// NewNonErrDoneGroup equals to WaitGroup. Only Todo() and WaitDone() func can be used!
func NewNonErrDoneGroup() NonErrDoneGroup {
	return &DoneGroup{sync.WaitGroup{}, nil, atomic.NewBool(false)}
}

func NewDoneGroup() DoneGroup {
	return DoneGroup{sync.WaitGroup{}, make(chan error, 1), atomic.NewBool(false)}
}

// Done equals to WaitGroup Done() but recover and call Error() on panic
func (d *DoneGroup) Done() {
	defer graceful.Recover(func(msg string) {
		d.Error(errors.New(msg))
	})
	d.WaitGroup.Done()
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
	if d.ec != nil {
		d.ec <- e
	}
}

func (d *DoneGroup) WaitError() <-chan error {
	return d.ec
}

func (d *DoneGroup) WaitDone() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		d.Wait()
		ch <- struct{}{}
	}()
	return ch
}

func (d *DoneGroup) Close() {
	if d.closed.CAS(false, true) {
		close(d.ec)
	}
}

//WaitUntilError use select to WaitDone() and WaitError() if has error return it else return nil
func (d *DoneGroup) WaitUntilError() error {
	for {
		select {
		case <-d.WaitDone():
			return nil
		case e := <-d.WaitError():
			return e
		}
	}
}
