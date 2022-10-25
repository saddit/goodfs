package util

import "testing"

func TestDoneGroupRecover(t *testing.T) {
	dg := NewDoneGroup()
	dg.Todo()
	go func () {
		defer dg.Done()
		var mp map[string]string
		mp["a"] = "a"
	}()
	t.Log(dg.WaitUntilError())
}