package set

type MapSet struct {
	mp map[any]struct{}
}

func (ms *MapSet) Add(elem any) {
	ms.mp[elem] = struct{}{}
}

func (ms *MapSet) Remove(elem any) bool {
	if ms.Contains(elem) {
		delete(ms.mp, elem)
		return true
	}
	return false
}

func (ms *MapSet) Contains(elem any) bool {
	_, ok := ms.mp[elem]
	return ok
}

func (ms *MapSet) Size() int {
	return len(ms.mp)
}

func (ms *MapSet) Foreach(fn func(elem any)) {
	for k := range ms.mp {
		fn(k)
	}
}

func NewMapSet() *MapSet {
	return &MapSet{mp: make(map[any]struct{})}
}
