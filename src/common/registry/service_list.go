package registry

import "sync"

type serviceList struct {
	data map[string]string // key=address value=registered key
	lock *sync.RWMutex
}

func newServiceList() *serviceList {
	return &serviceList{make(map[string]string), &sync.RWMutex{}}
}

func newServiceListOf(mp map[string]string) *serviceList {
	return &serviceList{mp, &sync.RWMutex{}}
}

func (s *serviceList) replace(mp map[string]string) {
	s.data = mp
}

func (s *serviceList) add(ip string, key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[ip] = key
}

func (s *serviceList) remove(ip string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.data, ip)
}

func (s *serviceList) list() []string {
	ls := make([]string, 0, len(s.data))
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k := range s.data {
		ls = append(ls, k)
	}
	return ls
}

func (s *serviceList) copy() map[string]string {
	ls := make(map[string]string, len(s.data))
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.data {
		ls[k] = v
	}
	return ls
}
