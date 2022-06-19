package util

import "sync"

type NonErrDoneGroup interface {
	Add(int)
	Done()
	Wait()
	WaitDone() <-chan bool
	Todo()
}

type DoneGroup struct {
	sync.WaitGroup
	ec chan error
}

func NewNonErrDoneGroup() NonErrDoneGroup {
	return &DoneGroup{sync.WaitGroup{}, nil}
}

func NewDoneGroup() DoneGroup {
	return DoneGroup{sync.WaitGroup{}, make(chan error, 1)}
}

//AddOne equals to wg.Add(1)
func (d *DoneGroup) Todo() {
	d.Add(1)
}

//Error deliver an error non blocking
func (d *DoneGroup) Error(e error) {
	d.ec <- e
}

func (d *DoneGroup) WaitError() <-chan error {
	return d.ec
}

func (d *DoneGroup) WaitDone() <-chan bool {
	ch := make(chan bool)
	go func() {
		d.Wait()
		ch <- true
	}()
	return ch
}

//Close close the error chan
func (d *DoneGroup) Close() {
	close(d.ec)
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
