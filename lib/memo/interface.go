package memo

import (
	"github.com/getevo/evo/v2/lib/serializer"
	"time"
)

const PERMANENT = time.Duration(-1)

type Interface interface {
	// Name returns driver name
	Name() string

	// Register initiate the driver
	Register() error

	// Set add an item to the cache, replacing any existing item. If the duration is 0
	Set(key string, value any, params ...any) error

	// SetRaw add an item to the cache, replacing any existing item. If the duration is 0
	SetRaw(key string, value []byte, params ...any) error

	// Replace set a new value for the cache key only if it already exists, and the existing
	// item hasn't expired. Returns an error otherwise.
	Replace(key string, value any, params ...any) bool

	// Get an item from the cache. Returns a bool indicating whether the key was found.
	Get(key string, out any, params ...any) bool

	// GetRaw get an item from the cache. Returns cache content in []byte and a bool indicating whether the key was found.
	GetRaw(key string, params ...any) ([]byte, bool)

	// GetWithExpiration returns an item and its expiration time from the cache.
	// It returns the item exported to out, the expiration time if one is set (if the item
	// never expires a zero value for time.Time is returned), and a bool indicating
	// whether the key was found.
	GetWithExpiration(key string, out any, params ...any) (time.Time, bool)

	// GetRawWithExpiration returns an item and its expiration time from the cache.
	// It returns the content in []byte, the expiration time if one is set (if the item
	// never expires a zero value for time.Time is returned), and a bool indicating
	// whether the key was found.
	GetRawWithExpiration(key string, params ...any) ([]byte, time.Time, bool)

	// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
	// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
	// item's value is not an integer, if it was not found, or if it is not
	// possible to increment it by n.
	Increment(key string, n any, params ...any) (int64, error)

	// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
	// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
	// item's value is not an integer, if it was not found, or if it is not
	// possible to decrement it by n.
	Decrement(key string, n any, params ...any) (int64, error)

	// Delete an item from the cache. Does nothing if the key is not in the cache.
	Delete(key string, params ...any) error

	// Expire re-set expiration duration for a key
	Expire(key string, t time.Time, params ...any) error

	// ItemCount Returns the number of items in the cache. This may include items that have
	// expired, but have not yet been cleaned up.
	ItemCount() int64

	// Flush delete all items from the cache.
	Flush() error

	SetSerializer(v serializer.Interface)

	Serializer() serializer.Interface

	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error

	// SetPrefix set a key prefix
	SetPrefix(p string)
}
