package util

import (
	"common/graceful"
	"errors"
	"fmt"
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

type DoneGroup interface {
	NonErrDoneGroup
	Close()
	Error(error)
	WaitError() <-chan error
	WaitUntilError() error
	ErrorUtilDone() <-chan error
}

type doneGroup struct {
	sync.WaitGroup
	ec     chan error
	closed *atomic.Bool
}

// NewNonErrDoneGroup equals to WaitGroup. Only Todo() and WaitDone() func can be used!
func NewNonErrDoneGroup() NonErrDoneGroup {
	return &doneGroup{sync.WaitGroup{}, nil, atomic.NewBool(true)}
}

func NewDoneGroup() DoneGroup {
	return &doneGroup{sync.WaitGroup{}, make(chan error, 1), atomic.NewBool(false)}
}

// Done equals to WaitGroup Done() but recover and call Error() on panic
func (d *doneGroup) Done() {
	// recover panic from d.WaitGroup.Done()
	defer func() {
		if err := recover(); err != nil {
			graceful.PrintStacks(err)
			d.Error(errors.New(fmt.Sprint(err)))
		}
	}()
	d.WaitGroup.Done()
	// recover panic of calling goroutine
	if err := recover(); err != nil {
		graceful.PrintStacks(err)
		d.Error(errors.New(fmt.Sprint(err)))
	}
}

// Todo equals to wg.Add(1)
func (d *doneGroup) Todo() {
	d.Add(1)
}

// Error deliver an error non blocking. Only one error can be received
func (d *doneGroup) Error(e error) {
	if d.closed.Load() {
		return
	}
	defer d.Close()
	if d.ec != nil {
		d.ec <- e
	}
}

func (d *doneGroup) WaitError() <-chan error {
	return d.ec
}

func (d *doneGroup) WaitDone() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer func() {
			if err := recover(); err != nil {
				graceful.PrintStacks(err)
				d.Error(errors.New(fmt.Sprint(err)))
			}
		}()
		defer close(ch)
		d.Wait()
		ch <- struct{}{}
	}()
	return ch
}

// Close closing the error receive channel. Safe to call multi goroutine and times
func (d *doneGroup) Close() {
	if d.closed.CAS(false, true) && d.ec != nil {
		close(d.ec)
	}
}

// WaitUntilError use select to WaitDone() and WaitError(). if has error return it else return nil
func (d *doneGroup) WaitUntilError() error {
	for {
		select {
		case <-d.WaitDone():
			return nil
		case e := <-d.WaitError():
			return e
		}
	}
}

// ErrorUtilDone receive errors until done.
func (d *doneGroup) ErrorUtilDone() <-chan error {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				graceful.PrintStacks(err)
				d.Error(errors.New(fmt.Sprint(err)))
			}
		}()
		defer d.Close()
		d.Wait()
	}()
	return d.ec
}

type limitDoneGroup struct {
	DoneGroup
	limitChan chan struct{}
}

// LimitDoneGroup limit only allows in Todo() and Done() function. Add() has not effect.
func LimitDoneGroup(max int) DoneGroup {
	return &limitDoneGroup{NewDoneGroup(), make(chan struct{}, max)}
}

func (ldg *limitDoneGroup) Todo() {
	ldg.limitChan <- struct{}{}
	ldg.DoneGroup.Todo()
}

func (ldg *limitDoneGroup) Done() {
	ldg.DoneGroup.Done()
	<-ldg.limitChan
	// recover panic of calling goroutine
	if err := recover(); err != nil {
		graceful.PrintStacks(err)
		ldg.Error(errors.New(fmt.Sprint(err)))
	}
}

func (ldg *limitDoneGroup) Close() {
	close(ldg.limitChan)
	ldg.DoneGroup.Close()
}
