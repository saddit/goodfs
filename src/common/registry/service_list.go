package registry

import "sync"

type serviceList struct {
	data map[string]string // key=address value=registered key
	lock *sync.RWMutex
}

func newServiceList() *serviceList {
	return &serviceList{make(map[string]string), &sync.RWMutex{}}
}

func (s serviceList) add(ip string, key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[ip] = key
}

func (s serviceList) remove(ip string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.data, ip)
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
