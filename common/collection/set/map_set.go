package set

type MapSet struct {
	mp map[interface{}]struct{}
}

func (ms *MapSet) Add(elem interface{}) {
	ms.mp[elem] = struct{}{}
}

func (ms *MapSet) Remove(elem interface{}) bool {
	if ms.Contains(elem) {
		delete(ms.mp, elem)
		return true
	}
	return false
}

func (ms *MapSet) Contains(elem interface{}) bool {
	_, ok := ms.mp[elem]
	return ok
}

func (ms *MapSet) Size() int {
	return len(ms.mp)
}

func (ms *MapSet) Foreach(fn func(elem interface{})) {
	for k := range ms.mp {
		fn(k)
	}
}

func NewMapSet() *MapSet {
	return &MapSet{mp: make(map[interface{}]struct{})}
}