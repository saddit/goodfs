package registry

import "sync"

type serviceList struct {
	data map[string]bool
	lock *sync.RWMutex
}

func newServiceList() *serviceList {
	return &serviceList{make(map[string]bool), &sync.RWMutex{}}
}

func (s serviceList) add(k string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[k] = true
}

func (s serviceList) remove(k string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.data, k)
}

func (s serviceList) list() []string {
	ls := make([]string, 0, len(s.data))
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k := range s.data {
		ls = append(ls, k)
	}
	return ls
}
