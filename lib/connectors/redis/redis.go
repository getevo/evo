package redis

import (
	"context"
	"sync"
	"time"

	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/memo/kv"
	"github.com/getevo/evo/v2/lib/pubsub"
	"github.com/getevo/evo/v2/lib/serializer"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/go-redis/redis/v8"
	"strings"
)

var Driver = &driver{
	listeners: map[string][]func(string, []byte, pubsub.Interface){},
	cancels:   map[string]context.CancelFunc{},
}

var _serializer = serializer.JSON
var prefix = ""
var Client redis.UniversalClient

type driver struct {
	mu        sync.RWMutex
	listeners map[string][]func(string, []byte, pubsub.Interface)
	// cancels holds a cancel func per topic to stop the subscriber goroutine.
	cancels map[string]context.CancelFunc
}

func (d *driver) Subscribe(topic string, onMessage func(topic string, message []byte, driver pubsub.Interface), params ...any) {
	topic = prefix + topic

	d.mu.Lock()
	d.listeners[topic] = append(d.listeners[topic], onMessage)
	alreadyRunning := d.cancels[topic] != nil
	var ctx context.Context
	if !alreadyRunning {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(context.Background())
		d.cancels[topic] = cancel
	}
	d.mu.Unlock()

	if alreadyRunning {
		return
	}

	go func() {
		for {
			sub := Client.Subscribe(ctx, topic)
			for {
				m, err := sub.ReceiveMessage(ctx)
				if err != nil {
					_ = sub.Close()
					if ctx.Err() != nil {
						// UnsubscribeAll was called — exit cleanly.
						return
					}
					log.Error("redis subscription error, reconnecting", "topic", topic, "error", err)
					time.Sleep(2 * time.Second)
					break
				}
				d.mu.RLock()
				callbacks := make([]func(string, []byte, pubsub.Interface), len(d.listeners[topic]))
				copy(callbacks, d.listeners[topic])
				d.mu.RUnlock()
				for _, callback := range callbacks {
					go callback(topic, []byte(m.Payload), d)
				}
			}
		}
	}()
}

// UnsubscribeAll removes all handlers for the given topic and stops its subscriber goroutine.
func (d *driver) UnsubscribeAll(topic string) {
	topic = prefix + topic
	d.mu.Lock()
	defer d.mu.Unlock()
	if cancel, ok := d.cancels[topic]; ok {
		cancel()
		delete(d.cancels, topic)
	}
	delete(d.listeners, topic)
}

func (d *driver) Publish(topic string, data any, params ...any) error {
	b, err := _serializer.Marshal(data)
	if err != nil {
		return err
	}
	return d.PublishBytes(topic, b, params...)
}

func (d *driver) PublishBytes(topic string, message []byte, params ...any) error {
	topic = prefix + topic
	return Client.Publish(context.Background(), topic, message).Err()
}

func (d *driver) Register() error {
	if Client != nil {
		return nil
	}
	_ = settings.Register(
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_ADDRESS",
			Title:       "Redis server(s) address",
			Description: "Redis server address. Separate multiple addresses with commas for cluster mode.",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_PREFIX",
			Title:       "Redis key prefix",
			Description: "Prefix for all keys to prevent collision when multiple apps share one Redis instance.",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_PASSWORD",
			Title:       "Redis password",
			Description: "Password for Redis authentication. Leave empty if not required.",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_DB",
			Title:       "Redis database index",
			Description: "Redis database number (0-15). Only applies to single-node mode. Default is 0.",
			Type:        "number",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_DIAL_TIMEOUT",
			Title:       "Redis dial timeout",
			Description: "Timeout for establishing a connection (e.g. 5s). Default is 5s.",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_READ_TIMEOUT",
			Title:       "Redis read timeout",
			Description: "Timeout for socket reads (e.g. 3s). Default is 3s.",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "CACHE",
			Name:        "REDIS_WRITE_TIMEOUT",
			Title:       "Redis write timeout",
			Description: "Timeout for socket writes (e.g. 3s). Default is 3s.",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
	)

	prefix = settings.Get("CACHE.REDIS_PREFIX").String()
	password := settings.Get("CACHE.REDIS_PASSWORD").String()
	db := settings.Get("CACHE.REDIS_DB", 0).Int()
	dialTimeout := settings.Get("CACHE.REDIS_DIAL_TIMEOUT", "5s").Duration()
	readTimeout := settings.Get("CACHE.REDIS_READ_TIMEOUT", "3s").Duration()
	writeTimeout := settings.Get("CACHE.REDIS_WRITE_TIMEOUT", "3s").Duration()

	rawAddrs := strings.Split(settings.Get("CACHE.REDIS_ADDRESS").String(), ",")
	var addrs []string
	for _, a := range rawAddrs {
		if trimmed := strings.TrimSpace(a); trimmed != "" {
			addrs = append(addrs, trimmed)
		}
	}

	if len(addrs) == 1 {
		Client = redis.NewClient(&redis.Options{
			Addr:         addrs[0],
			Password:     password,
			DB:           db,
			DialTimeout:  dialTimeout,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		})
	} else {
		Client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        addrs,
			Password:     password,
			DialTimeout:  dialTimeout,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		})
	}
	return Client.Ping(context.Background()).Err()
}

func (d *driver) Name() string {
	return "redis"
}

// Set adds an item to the cache, replacing any existing item.
func (d *driver) Set(key string, value any, params ...any) error {
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

// SetRaw adds raw bytes to the cache, replacing any existing item.
func (d *driver) SetRaw(key string, value []byte, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	return Client.Set(context.Background(), prefix+key, value, p.Duration).Err()
}

// SetNX sets key to value only if the key does not already exist.
// Returns true if the key was set, false if it already existed.
func (d *driver) SetNX(key string, value any, params ...any) (bool, error) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	b, err := _serializer.Marshal(value)
	if err != nil {
		return false, err
	}
	return Client.SetNX(context.Background(), prefix+key, b, p.Duration).Result()
}

// Replace sets a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (d *driver) Replace(key string, value any, params ...any) bool {
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

// Get retrieves an item from the cache. Returns false if the key was not found.
func (d *driver) Get(key string, out any, params ...any) bool {
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
		log.Error("unable to unmarshal redis value", "error", err)
		return false
	}
	return true
}

// GetRaw retrieves raw bytes from the cache. Returns nil and false if the key was not found.
func (d *driver) GetRaw(key string, params ...any) ([]byte, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	result, err := Client.Get(context.Background(), prefix+key).Bytes()
	if err != nil {
		return nil, false
	}
	return result, true
}

// GetWithExpiration returns an item and its expiration time from the cache.
// A zero expiration time means the key has no expiry.
// Returns false if the key was not found.
func (d *driver) GetWithExpiration(key string, out any, params ...any) (time.Time, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	if !d.Get(key, out) {
		return time.Time{}, false
	}
	ttl, err := Client.TTL(context.Background(), prefix+key).Result()
	if err != nil {
		return time.Time{}, false
	}
	if ttl < 0 {
		// Key exists but has no expiry.
		return time.Time{}, true
	}
	return time.Now().Add(ttl), true
}

// GetRawWithExpiration returns raw bytes and expiration time from the cache.
// A zero expiration time means the key has no expiry.
// Returns false if the key was not found.
func (d *driver) GetRawWithExpiration(key string, params ...any) ([]byte, time.Time, bool) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	result, err := Client.Get(context.Background(), prefix+key).Bytes()
	if err != nil {
		return nil, time.Time{}, false
	}
	ttl, err := Client.TTL(context.Background(), prefix+key).Result()
	if err != nil {
		return nil, time.Time{}, false
	}
	if ttl < 0 {
		// Key exists but has no expiry.
		return result, time.Time{}, true
	}
	return result, time.Now().Add(ttl), true
}

// Exists reports whether the given key is present in the cache.
func (d *driver) Exists(key string, params ...any) bool {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	n, err := Client.Exists(context.Background(), prefix+key).Result()
	return err == nil && n > 0
}

// Keys returns all keys matching the given glob pattern using Redis SCAN (non-blocking).
// Use "*" to list all keys or "prefix:*" to match a namespace.
// Note: in cluster mode only the keys on the connected node are returned.
func (d *driver) Keys(pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := Client.Scan(context.Background(), cursor, prefix+pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// Increment atomically increments the integer value of a key by n.
// Returns an error if the key's value is not an integer or does not exist.
func (d *driver) Increment(key string, n any, params ...any) (int64, error) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v := toInt64(n)
	result := Client.IncrBy(context.Background(), prefix+key, v)
	return result.Val(), result.Err()
}

// Decrement atomically decrements the integer value of a key by n.
// Returns an error if the key's value is not an integer or does not exist.
func (d *driver) Decrement(key string, n any, params ...any) (int64, error) {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	v := toInt64(n)
	result := Client.DecrBy(context.Background(), prefix+key, v)
	return result.Val(), result.Err()
}

// Delete removes an item from the cache. Does nothing if the key does not exist.
func (d *driver) Delete(key string, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	return Client.Del(context.Background(), prefix+key).Err()
}

// Expire resets the expiration time of a key to the given absolute time.
func (d *driver) Expire(key string, t time.Time, params ...any) error {
	var p = kv.Parse(params)
	if p.Bucket != "" {
		key = p.Bucket + "." + key
	}
	return Client.Expire(context.Background(), prefix+key, time.Until(t)).Err()
}

// ItemCount returns the number of keys in the current database.
// In cluster mode this reflects only the connected node's count.
func (d *driver) ItemCount() int64 {
	return Client.DBSize(context.Background()).Val()
}

// Flush deletes all keys in the current database.
// In cluster mode this only affects the connected node.
func (d *driver) Flush() error {
	return Client.FlushDB(context.Background()).Err()
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
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	}
	return 0
}

// SetSerializer sets the data serialization method.
func (d *driver) SetSerializer(v serializer.Interface) {
	_serializer = v
}

func (d *driver) SetPrefix(s string) {
	prefix = s
}

func (d *driver) Serializer() serializer.Interface {
	return _serializer
}

func (d *driver) Marshal(data any) ([]byte, error) {
	return _serializer.Marshal(data)
}

func (d *driver) Unmarshal(data []byte, v any) error {
	return _serializer.Unmarshal(data, v)
}
