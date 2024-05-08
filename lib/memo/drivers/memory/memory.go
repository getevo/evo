package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/memo/kv"
	"github.com/getevo/evo/v2/lib/serializer"
	"github.com/getevo/evo/v2/lib/settings"
)

var Driver = driver{}

type driver struct{}

var _serializer serializer.Interface

func (d driver) SetPrefix(p string) {
	return
}

var items sync.Map

func (driver) Register() error {
	items = sync.Map{}

	settings.Register(
		settings.SettingDomain{
			Title:       "Cache",
			Domain:      "CACHE",
			Description: "system cache configurations",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "MEMORY_JANITOR_INTERVAL",
			Title:       "IN-Memory cache janitor interval",
			Description: "The interval which janitor start to clean the memory for evicted items.",
			Type:        "duration",
			Value:       "1m",
			ReadOnly:    false,
			Visible:     true,
		})
	var sleep, err = settings.Get("CACHE.MEMORY_JANITOR_INTERVAL").Duration()
	if err != nil || sleep < 1*time.Second {
		log.Warning("Invalid CACHE.MEMORY_JANITOR_INTERVAL", "value", settings.Get("CACHE.MEMORY_JANITOR_INTERVAL").String())
		sleep = 1 * time.Minute
	}
	go func() {
		for {
			time.Sleep(sleep)
			var now = time.Now().Unix()
			items.Range(func(key, value any) bool {
				fmt.Printf("%+v", value)
				if value.(item).expires < now {
					items.Delete(key)
				}
				return true
			})
		}
	}()
	return nil
}

// An item represents arbitrary data with expiration time.
type item struct {
	data    []byte
	expires int64
}

func (driver) Name() string {
	return "memory"
}

// Set add an item to the cache, replacing any existing item. If the duration is 0
func (driver) Set(key string, value any, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	b, err := _serializer.Marshal(value)
	if err != nil {
		return err
	}
	items.Store(key, item{
		data:    b,
		expires: time.Now().Add(p.Duration).Unix(),
	})
	return nil

}

// SetRaw add an item to the cache, replacing any existing item. If the duration is 0
func (driver) SetRaw(key string, value []byte, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	items.Store(key, item{
		data:    value,
		expires: time.Now().Add(p.Duration).Unix(),
	})
	return nil
}

// Replace set a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (d driver) Replace(key string, value any, params ...any) bool {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	if _, ok := d.GetRaw(key); !ok {
		return false
	}
	d.Set(key, value, p.Duration)
	return true
}

// Get an item from the cache. Returns a bool indicating whether the key was found.
func (driver) Get(key string, out any, params ...any) bool {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v, ok := items.Load(key)
	if !ok {
		return false
	}
	if v.(item).expires < time.Now().Unix() {
		return false
	}
	var err = _serializer.Unmarshal(v.(item).data, out)
	if err != nil {
		log.Error("unable to unmarshal message", "error", err)
		return false
	}
	return true
}

// GetRaw get an item from the cache. Returns cache content in []byte and a bool indicating whether the key was found.
func (driver) GetRaw(key string, params ...any) ([]byte, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v, ok := items.Load(key)
	if !ok {
		return nil, false
	}
	if v.(item).expires < time.Now().Unix() {
		return nil, false
	}
	return v.(item).data, false
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item exported to out, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (d driver) GetWithExpiration(key string, out any, params ...any) (time.Time, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v, ok := items.Load(key)
	if !ok {
		return time.Time{}, false
	}
	if v.(item).expires < time.Now().Unix() {
		return time.Time{}, false
	}
	var err = _serializer.Unmarshal(v.(item).data, out)
	if err != nil {
		log.Error("unable to unmarshal message", "error", err)
		return time.Time{}, false
	}
	return time.Unix(v.(item).expires, 0), true
}

// GetRawWithExpiration returns an item and its expiration time from the cache.
// It returns the content in []byte, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (driver) GetRawWithExpiration(key string, params ...any) ([]byte, time.Time, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v, ok := items.Load(key)
	if !ok {
		return nil, time.Time{}, false
	}
	if v.(item).expires < time.Now().Unix() {
		return nil, time.Time{}, false
	}
	return v.(item).data, time.Unix(v.(item).expires, 0), true
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n.
func (d driver) Increment(key string, n any, params ...any) (int64, error) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	var v int64

	loaded, ok := items.Load(key)

	var i item
	if ok && loaded.(item).expires >= time.Now().Unix() {
		i = loaded.(item)
		var err = _serializer.Unmarshal(i.data, &v)
		if err != nil {
			return 0, err
		}
		v = v + toInt64(n)
		i.data, _ = _serializer.Marshal(v)
	} else {
		i = item{
			expires: -1,
		}
		i.data, _ = _serializer.Marshal(1)
	}
	items.Store(key, i)
	return v, nil
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n.
func (driver) Decrement(key string, n any, params ...any) (int64, error) {
	var v int64
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	loaded, ok := items.Load(key)

	var i item
	if ok && loaded.(item).expires >= time.Now().Unix() {
		i = loaded.(item)
		var err = _serializer.Unmarshal(i.data, &v)
		if err != nil {
			return 0, err
		}
		v = v - toInt64(n)
		i.data, _ = _serializer.Marshal(v)
	} else {
		i = item{
			expires: -1,
		}
		i.data, _ = _serializer.Marshal(1)
	}
	items.Store(key, i)
	return v, nil
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (driver) Delete(key string, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	items.Delete(key)
	return nil
}

// Expire re-set expiration duration for a key
func (driver) Expire(key string, t time.Time, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	loaded, ok := items.Load(key)
	if !ok {
		return fmt.Errorf("key not found")
	}
	var i = loaded.(item)
	i.expires = t.Unix()
	items.Store(key, i)
	return nil
}

// ItemCount Returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (driver) ItemCount() int64 {
	return 0
}

// Flush delete all items from the cache.
func (driver) Flush() error {
	items = sync.Map{}
	return nil
}

func toInt64(n any) int64 {
	switch v := n.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	}
	return 0
}

func (driver) SetSerializer(v serializer.Interface) {
	_serializer = v
}
func (driver) Serializer() serializer.Interface {
	return _serializer
}

func (driver) Marshal(data any) ([]byte, error) {
	return _serializer.Marshal(data)
}

func (driver) Unmarshal(data []byte, v any) error {
	return _serializer.Unmarshal(data, v)
}
