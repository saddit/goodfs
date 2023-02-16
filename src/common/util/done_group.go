package util

import (
	"common/graceful"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type WaitGroup interface {
	Add(int)
	Done()
	Wait()
	WaitDone() <-chan struct{}
	Todo()
}

type DoneGroup interface {
	WaitGroup
	// Close stop receive error
	Close()
	// Error deliver an error and close the group. Only one error can be received.
	Error(error)
	// Errors deliver an error. Must call WaitError first in another goroutine to prevent from deadlock.
	// Do not call Error or Close without waiting the group done if this function was used in any goroutine
	Errors(error)
	// WaitError block to receive err from Error or Errors until group is closed
	WaitError() <-chan error
	// WaitUntilError use select to WaitDone() and WaitError(). if has error return it else return nil
	WaitUntilError() error
}

type doneGroup struct {
	sync.WaitGroup
	ec     chan error
	closed *atomic.Bool
}

func NewWaitGroup() WaitGroup {
	b := &atomic.Bool{}
	b.Store(true)
	return &doneGroup{sync.WaitGroup{}, nil, b}
}

func NewDoneGroup() DoneGroup {
	return &doneGroup{sync.WaitGroup{}, make(chan error, 1), &atomic.Bool{}}
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

// Error deliver an error and close the group. Only one error can be received.
func (d *doneGroup) Error(e error) {
	if d.closed.CompareAndSwap(false, true) && d.ec != nil {
		defer close(d.ec)
		d.ec <- e
	}
}

// Errors deliver an error. Must call WaitError first in another goroutine to prevent from deadlock.
// Do not call Error or Close without waiting the group done if this function was used in any goroutine
func (d *doneGroup) Errors(e error) {
	if d.closed.Load() {
		return
	}
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
	if d.closed.CompareAndSwap(false, true) && d.ec != nil {
		close(d.ec)
	}
}

// WaitUntilError use select to WaitDone() and WaitError(). if has error return it else return nil
func (d *doneGroup) WaitUntilError() error {
	select {
	case <-d.WaitDone():
		return nil
	case e := <-d.WaitError():
		return e
	}
}

type limitDoneGroup struct {
	DoneGroup
	limitChan chan struct{}
}

// LimitDoneGroup limit only allows in Todo() and Done() function. Add() has not effect.
func LimitDoneGroup(max int) *limitDoneGroup {
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
