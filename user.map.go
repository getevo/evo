package evo

import "sync"

type cMap struct {
	lock sync.Mutex
	data map[uint]interface{}
}

// Init concurrent map
func (m *cMap) Init() {
	m.data = map[uint]interface{}{}
}

// Set concurrent map item
func (m *cMap) Set(k uint, v interface{}) {
	m.lock.Lock()
	m.data[k] = v
	m.lock.Unlock()
}

// Get concurrent map item
func (m *cMap) Get(k uint) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	if v, ok := m.data[k]; ok {
		return v
	}
	return nil
}

// Has check if concurrent map has key
func (m *cMap) Has(k uint) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.data[k]; ok {
		return true
	}
	return false
}

// Count return number of items in concurrent map
func (m *cMap) Count() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.data)
}
