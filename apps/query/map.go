package query

import "sync"

type Map struct {
	lock sync.Mutex
	data map[string]Filter
}

func (m *Map) Init() {
	m.data = map[string]Filter{}
}

func (m *Map) Set(k string, v Filter) {
	m.lock.Lock()
	m.data[k] = v
	m.lock.Unlock()
}

func (m *Map) Get(k string) *Filter {
	m.lock.Lock()
	defer m.lock.Unlock()
	if v, ok := m.data[k]; ok {
		return &v
	}
	return nil
}

func (m *Map) Has(k string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.data[k]; ok {
		return true
	}
	return false
}

func (m *Map) Count() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.data)
}
