package cache

import (
	"time"
)

const PERMANENT = -1

type Interface interface {
	// Name returns driver name
	Name() string

	// Register initiate the driver
	Register() error

	// Set add an item to the cache, replacing any existing item. If the duration is 0
	Set(key string, value interface{}, duration time.Duration)

	// SetRaw add an item to the cache, replacing any existing item. If the duration is 0
	SetRaw(key string, value []byte, duration time.Duration)

	// Replace set a new value for the cache key only if it already exists, and the existing
	// item hasn't expired. Returns an error otherwise.
	Replace(key string, value interface{}, duration time.Duration) bool

	// Get an item from the cache. Returns a bool indicating whether the key was found.
	Get(key string, out interface{}) bool

	// GetRaw get an item from the cache. Returns cache content in []byte and a bool indicating whether the key was found.
	GetRaw(key string) ([]byte, bool)

	// GetWithExpiration returns an item and its expiration time from the cache.
	// It returns the item exported to out, the expiration time if one is set (if the item
	// never expires a zero value for time.Time is returned), and a bool indicating
	// whether the key was found.
	GetWithExpiration(key string, out interface{}) (time.Time, bool)

	// GetRawWithExpiration returns an item and its expiration time from the cache.
	// It returns the content in []byte, the expiration time if one is set (if the item
	// never expires a zero value for time.Time is returned), and a bool indicating
	// whether the key was found.
	GetRawWithExpiration(key string) ([]byte, time.Time, bool)

	// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
	// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
	// item's value is not an integer, if it was not found, or if it is not
	// possible to increment it by n.
	Increment(key string, n interface{}) int64

	// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
	// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
	// item's value is not an integer, if it was not found, or if it is not
	// possible to decrement it by n.
	Decrement(key string, n interface{}) int64

	// Delete an item from the cache. Does nothing if the key is not in the cache.
	Delete(key string)

	// Expire re-set expiration duration for a key
	Expire(key string, t time.Time) error

	// ItemCount Returns the number of items in the cache. This may include items that have
	// expired, but have not yet been cleaned up.
	ItemCount() int64

	// Flush delete all items from the cache.
	Flush() error

	// SetMarshaller set interface{} to []byte marshalling function
	SetMarshaller(func(input interface{}) ([]byte, error))

	// SetUnMarshaller set []byte to interface{} unmarshalling function
	SetUnMarshaller(func(bytes []byte, out interface{}) error)
}
