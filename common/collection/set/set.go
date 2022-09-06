package set

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
