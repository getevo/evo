package concurrent

import (
	"sort"
	"sync"
)

// Map concurrent string to interface map
type Map struct {
	lock sync.Mutex
	data map[string]interface{}
}

// Init initializes a new map
func (m *Map) Init() {
	m.lock.Lock()
	m.data = map[string]interface{}{}
	m.lock.Unlock()
}

// Set sets the map value for given key
func (m *Map) Set(k string, v interface{}) {
	m.lock.Lock()
	m.data[k] = v
	m.lock.Unlock()
}

// Get gets map value for given key. return nil if not found
func (m *Map) Get(k string) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	if v, ok := m.data[k]; ok {
		return v
	}
	return nil
}

// Has check if map has key
func (m *Map) Has(k string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.data[k]; ok {
		return true
	}
	return false
}

// Count return number of map items
func (m *Map) Count() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.data)
}

// Data return direct pointer to map
func (m *Map) Data() map[string]interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.data
}

// Keys return sorted map keys
func (m *Map) Keys() []string {
	m.lock.Lock()
	defer m.lock.Unlock()
	var list []string
	for k := range m.data {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}
