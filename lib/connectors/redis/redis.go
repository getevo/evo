package redis

import (
	"context"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/memo/kv"
	"github.com/getevo/evo/v2/lib/pubsub"
	"github.com/getevo/evo/v2/lib/serializer"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
)

var Driver = driver{}
var _serializer = serializer.JSON
var listeners = map[string][]func(topic string, message []byte, driver pubsub.Interface){}

type driver struct{}

func (d driver) Subscribe(topic string, onMessage func(topic string, message []byte, driver pubsub.Interface), params ...any) {
	topic = prefix + topic
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

func (d driver) Publish(topic string, data any, params ...any) error {
	b, err := _serializer.Marshal(data)
	if err != nil {
		return err
	}
	return d.PublishBytes(topic, b, params...)
}

func (d driver) PublishBytes(topic string, message []byte, params ...any) error {
	topic = prefix + topic
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
func (driver) Set(key string, value any, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	b, err := _serializer.Marshal(value)
	if err != nil {
		return err
	}
	return Client.Set(context.Background(), prefix+key, b, p.Duration).Err()
}

// SetRaw add an item to the cache, replacing any existing item. If the duration is 0
func (driver) SetRaw(key string, value []byte, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	return Client.Set(context.Background(), prefix+key, value, p.Duration).Err()
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
	err := d.Set(key, value, p.Duration)
	if err != nil {
		return false
	}
	return true
}

// Get an item from the cache. Returns a bool indicating whether the key was found.
func (driver) Get(key string, out any, params ...any) bool {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	result, err := Client.Get(context.Background(), prefix+key).Bytes()
	if err != nil {
		return false
	}
	err = _serializer.Unmarshal(result, out)
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
func (d driver) GetWithExpiration(key string, out any, params ...any) (time.Time, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
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
func (driver) GetRawWithExpiration(key string, params ...any) ([]byte, time.Time, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
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
func (driver) Increment(key string, n any, params ...any) (int64, error) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v := toInt64(n)
	var result = Client.IncrBy(context.Background(), prefix+key, v)
	return result.Val(), result.Err()
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n.
func (driver) Decrement(key string, n any, params ...any) (int64, error) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v := toInt64(n)
	var result = Client.IncrBy(context.Background(), prefix+key, v)
	return result.Val(), result.Err()
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (driver) Delete(key string, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	return Client.Del(context.Background(), prefix+key).Err()
}

// Expire re-set expiration duration for a key
func (driver) Expire(key string, t time.Time, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
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

// SetSerializer set data serialization method
func (driver) SetSerializer(v serializer.Interface) {
	_serializer = v
}

func (driver) SetPrefix(s string) {
	prefix = s
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
