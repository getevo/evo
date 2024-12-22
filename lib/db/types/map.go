package types

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"sync"
)

const SHARD_COUNT = 32

// Stringer is an interface that combines fmt.Stringer and comparable.
// It ensures that the type can be used both as a string and as a map key.
type Stringer interface {
	fmt.Stringer
	comparable
}

// Map represents a thread-safe map using sharded storage.
// It divides the map into multiple shards to reduce lock contention and improve performance.
type Map[K comparable, V any] struct {
	shards   []*ConcurrentMapShared[K, V]
	sharding func(key K) uint32
}

// ConcurrentMapShared represents a single shard of the ConcurrentMap.
// It contains a map guarded by a RWMutex for safe concurrent access.
type ConcurrentMapShared[K comparable, V any] struct {
	items        map[K]V
	sync.RWMutex // Protects access to the internal map.
}

// create initializes a ConcurrentMap with a custom sharding function.
func create[K comparable, V any](sharding func(key K) uint32) Map[K, V] {
	m := Map[K, V]{
		sharding: sharding,
		shards:   make([]*ConcurrentMapShared[K, V], SHARD_COUNT),
	}
	for i := 0; i < SHARD_COUNT; i++ {
		m.shards[i] = &ConcurrentMapShared[K, V]{items: make(map[K]V)}
	}
	return m
}

// NewMap creates a new ConcurrentMap with string keys and a default sharding function.
func NewMap[V any]() Map[string, V] {
	return create[string, V](fnv32)
}

// NewStringerMap creates a new Map with custom Stringer keys and a sharding function.
func NewStringerMap[K Stringer, V any]() Map[K, V] {
	return create[K, V](strfnv32[K])
}

// NewMapWithCustomShardingFunction creates a new Map with a user-defined sharding function.
func NewMapWithCustomShardingFunction[K comparable, V any](sharding func(key K) uint32) Map[K, V] {
	return create[K, V](sharding)
}

// GetShard returns the shard corresponding to the given key.
func (m Map[K, V]) GetShard(key K) *ConcurrentMapShared[K, V] {
	return m.shards[uint(m.sharding(key))%uint(SHARD_COUNT)]
}

// MSet sets multiple key-value pairs in the map concurrently.
func (m Map[K, V]) MSet(data map[K]V) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

// Set inserts or updates a value in the map under the specified key.
func (m Map[K, V]) Set(key K, value V) {
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// UpsertCb defines a callback function for Upsert operations.
// The callback receives whether the key exists, the current value, and the new value.
type UpsertCb[V any] func(exist bool, valueInMap V, newValue V) V

// Upsert inserts or updates a value in the map using a callback function.
func (m Map[K, V]) Upsert(key K, value V, cb UpsertCb[V]) (res V) {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// SetIfAbsent sets a value in the map only if the key does not already exist.
func (m Map[K, V]) SetIfAbsent(key K, value V) bool {
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

// Get retrieves a value by key from the map.
func (m Map[K, V]) Get(key K) (V, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Remove deletes a key-value pair from the map.
func (m Map[K, V]) Remove(key K) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// RemoveCb defines a callback for conditional removal.
type RemoveCb[K any, V any] func(key K, v V, exists bool) bool

// RemoveCb removes a key-value pair if the callback condition returns true.
func (m Map[K, V]) RemoveCb(key K, cb RemoveCb[K, V]) bool {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

// fnv32 is a hashing function for string keys.
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// strfnv32 is a hashing function for Stringer keys.
func strfnv32[K fmt.Stringer](key K) uint32 {
	return fnv32(key.String())
}

// GormDBDataType defines how GORM should handle the type in different databases.
func (m *Map[K, V]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return "TEXT"
}
