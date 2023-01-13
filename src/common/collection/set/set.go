package set

import "reflect"

type Set interface {
	Contains(elem any) bool
	Add(elem any)
	Remove(elem any) bool
	Size() int
	Foreach(fn func(elem any))
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

func To[T any](st Set) []T {
	var arr []T
	st.Foreach(func(elem any) {
		arr = append(arr, elem.(T))
	})
	return arr
}
