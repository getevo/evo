package redis

import (
	"context"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/pubsub"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/go-redis/redis/v8"
	"github.com/kelindar/binary"
	"strings"
	"time"
)

var Driver = driver{}
var marshaller func(input interface{}) ([]byte, error) = binary.Marshal
var unmarshaller func(bytes []byte, out interface{}) error = binary.Unmarshal
var listeners = map[string][]func(topic string, message []byte, driver pubsub.Interface){}

type driver struct{}

func (d driver) Subscribe(topic string, onMessage func(topic string, message []byte, driver pubsub.Interface), params ...interface{}) {
	if _, ok := listeners[topic]; !ok {
		listeners[topic] = []func(topic string, message []byte, driver pubsub.Interface){}
	}
	listeners[topic] = append(listeners[topic], onMessage)
	go func() {
		pubsub := Client.Subscribe(context.Background(), prefix+topic)
		for {
			m, err := pubsub.ReceiveMessage(context.Background())
			if err != nil {
				log.Error("unable to receive redis message", "error", err)
			}
			for _, callback := range listeners[topic] {
				go callback(topic, []byte(m.Payload), d)
			}

		}
	}()
}
func (d driver) Publish(topic string, message []byte, params ...interface{}) error {
	return Client.Publish(context.Background(), prefix+topic, message).Err()
}

var prefix = ""
var Client redis.UniversalClient

func (driver) Register() error {
	if Client != nil {
		return nil
	}
	settings.Register(settings.Setting{
		Domain:      "CACHE",
		Name:        "REDIS_ADDRESS",
		Title:       "Redis server(s) address",
		Description: "Redis servers address. separate using comma if cluster.",
		Type:        "text",
		ReadOnly:    false,
		Visible:     true,
	},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_PREFIX",
			Title:       "Redis key prefix",
			Description: "Set a prefix for keys to prevent conjunction of keys in case of multi application running on same instance of redis",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
	)
	prefix = settings.Get("CACHE.REDIS_PREFIX").String()
	var addrs = strings.Split(settings.Get("CACHE.REDIS_ADDRESS").String(), ",")

	if len(addrs) == 1 {
		Client = redis.NewClient(&redis.Options{
			Addr: addrs[0],
		})
	} else {
		Client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: addrs,
		})
	}
	return Client.Ping(context.Background()).Err()
}

func (driver) Name() string {
	return "redis"
}

// Set add an item to the cache, replacing any existing item. If the duration is 0
func (driver) Set(key string, value interface{}, duration time.Duration) {
	b, err := marshaller(value)
	if err != nil {
		log.Error("unable to marshal message", "error", err)
		return
	}
	Client.Set(context.Background(), prefix+key, b, duration)
}

// SetRaw add an item to the cache, replacing any existing item. If the duration is 0
func (driver) SetRaw(key string, value []byte, duration time.Duration) {
	Client.Set(context.Background(), prefix+key, value, duration)
}

// Replace set a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (d driver) Replace(key string, value interface{}, duration time.Duration) bool {
	if _, ok := d.GetRaw(key); !ok {
		return false
	}
	d.Set(key, value, duration)
	return true
}

// Get an item from the cache. Returns a bool indicating whether the key was found.
func (driver) Get(key string, out interface{}) bool {
	result, err := Client.Get(context.Background(), prefix+key).Bytes()
	if err != nil {
		return false
	}
	err = unmarshaller(result, out)
	if err != nil {
		log.Error("unable to unmarshal message", "error", err)
		return false
	}
	return true
}

// GetRaw get an item from the cache. Returns cache content in []byte and a bool indicating whether the key was found.
func (driver) GetRaw(key string) ([]byte, bool) {
	result, err := Client.Get(context.Background(), prefix+key).Bytes()
	if err == nil {
		return result, true
	}
	return result, false
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item exported to out, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (d driver) GetWithExpiration(key string, out interface{}) (time.Time, bool) {
	var exists = d.Get(key, out)
	if !exists {
		return time.Time{}, false
	}
	ttl, err := Client.TTL(context.Background(), prefix+key).Result()
	if err != nil {
		return time.Time{}, false
	}
	return time.Now().Add(ttl), true
}

// GetRawWithExpiration returns an item and its expiration time from the cache.
// It returns the content in []byte, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (driver) GetRawWithExpiration(key string) ([]byte, time.Time, bool) {
	result, err := Client.Get(context.Background(), prefix+key).Bytes()
	if err != nil {
		return result, time.Time{}, false
	}
	ttl, err := Client.TTL(context.Background(), prefix+key).Result()
	if err != nil {
		return result, time.Time{}, false
	}
	return result, time.Now().Add(ttl), true
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n.
func (driver) Increment(key string, n interface{}) int64 {
	v := toInt64(n)
	var result = Client.IncrBy(context.Background(), prefix+key, v)
	return result.Val()
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n.
func (driver) Decrement(key string, n interface{}) int64 {
	v := toInt64(n)
	var result = Client.IncrBy(context.Background(), prefix+key, v)
	return result.Val()
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (driver) Delete(key string) {
	Client.Del(context.Background(), prefix+key)
}

// Expire re-set expiration duration for a key
func (driver) Expire(key string, t time.Time) error {
	return Client.Expire(context.Background(), prefix+key, time.Now().Sub(t)).Err()
}

// ItemCount Returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (driver) ItemCount() int64 {
	return 0
}

// Flush delete all items from the cache.
func (driver) Flush() error {
	return nil
}

// SetMarshaller set interface{} to []byte marshalling function
func (driver) SetMarshaller(fn func(input interface{}) ([]byte, error)) {
	marshaller = fn
}

// SetUnMarshaller set []byte to interface{} unmarshalling function
func (driver) SetUnMarshaller(fn func(bytes []byte, out interface{}) error) {
	unmarshaller = fn
}

func toInt64(n interface{}) int64 {
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
