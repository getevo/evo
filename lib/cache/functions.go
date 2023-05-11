package cache

import (
	"github.com/getevo/evo/v2/lib/cache/drivers/memory"
	"github.com/getevo/evo/v2/lib/log"
	"time"
)

var defaultDriver Interface = nil
var drivers = map[string]Interface{}

func SetDefaultDriver(driver Interface) {
	AddDriver(driver)
	defaultDriver = driver
}

func DriverName() string {
	return defaultDriver.Name()
}

func Drivers() map[string]Interface {
	return drivers
}

func Driver(driver string) (Interface, bool) {
	if v, ok := drivers[driver]; ok {
		return v, ok
	}
	return nil, false
}

func Use(driver string) Interface {
	return drivers[driver]
}

func AddDriver(driver Interface) {
	if _, ok := drivers[driver.Name()]; !ok {
		drivers[driver.Name()] = driver
		var err = drivers[driver.Name()].Register()
		if err != nil {
			log.Fatal("unable to initiate cache driver", "name", driver.Name(), "error", err)
		}
	}
	if defaultDriver == nil {
		defaultDriver = driver
	}
}

// Set add an item to the cache, replacing any existing item. If the duration is 0
func Set(key string, value interface{}, duration time.Duration) {
	defaultDriver.Set(key, value, duration)
}

// SetRaw add an item to the cache, replacing any existing item. If the duration is 0
func SetRaw(key string, value []byte, duration time.Duration) {
	defaultDriver.SetRaw(key, value, duration)
}

// Replace set a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func Replace(key string, value interface{}, duration time.Duration) bool {
	return defaultDriver.Replace(key, value, duration)
}

// Get an item from the cache. Returns a bool indicating whether the key was found.
func Get(key string, out interface{}) bool {
	return defaultDriver.Get(key, out)
}

// GetRaw get an item from the cache. Returns cache content in []byte and a bool indicating whether the key was found.
func GetRaw(key string) ([]byte, bool) {
	return defaultDriver.GetRaw(key)
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item exported to out, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func GetWithExpiration(key string, out interface{}) (time.Time, bool) {
	return defaultDriver.GetWithExpiration(key, out)
}

// GetRawWithExpiration returns an item and its expiration time from the cache.
// It returns the content in []byte, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func GetRawWithExpiration(key string) ([]byte, time.Time, bool) {
	return defaultDriver.GetRawWithExpiration(key)
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n.
func Increment(key string, n interface{}) int64 {
	return defaultDriver.Increment(key, n)
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n.
func Decrement(key string, n interface{}) int64 {
	return defaultDriver.Decrement(key, n)
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func Delete(key string) {
	defaultDriver.Delete(key)
}

// Expire set expiration date for a key
func Expire(key string, t time.Time) error {
	return defaultDriver.Expire(key, t)
}

// ItemCount Returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func ItemCount() int64 {
	return defaultDriver.ItemCount()
}

// Flush delete all items from the cache.
func Flush() error {
	return defaultDriver.Flush()
}

func Register() error {
	SetDefaultDriver(memory.Driver)
	return nil
}

func DefaultDriver() Interface {
	return defaultDriver
}
