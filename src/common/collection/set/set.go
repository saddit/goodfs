package set

import "reflect"

type Set interface {
	Contains(elem interface{}) bool
	Add(elem interface{})
	Remove(elem interface{}) bool
	Size() int
	Foreach(fn func(elem interface{}))
}

func OfInteger(arr []int) Set {
	mpSet := NewMapSet()
	for _, v := range arr {
		mpSet.Add(v)
	}
	return mpSet
}

func OfString(arr []string) Set {
	mpSet := NewMapSet()
	for _, v := range arr {
		mpSet.Add(v)
	}
	return mpSet
}

func OfMapKeys(mp any) Set {
	mpVal := reflect.Indirect(reflect.ValueOf(mp))
	st := NewMapSet()
	for _, k := range mpVal.MapKeys() {
		if k.CanInterface() {
			st.Add(k.Interface())
		}
	}
	return st
}
