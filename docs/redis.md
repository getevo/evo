# Redis Connector

EVO's Redis connector (`lib/connectors/redis`) provides a unified cache, key-value store, and pub/sub driver backed by Redis. It supports both single-node and cluster deployments.

## Setup

Register the driver during application startup:

```go
import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/connectors/redis"
    "github.com/getevo/evo/v2/lib/pubsub"
)

func main() {
    evo.Setup(redis.Driver)

    // Optionally set as the default pub/sub driver
    pubsub.SetDefaultDriver(redis.Driver)

    evo.Run()
}
```

## Configuration

All settings are registered automatically under the `CACHE` domain.

| Key                       | Type   | Default | Description |
|---------------------------|--------|---------|-------------|
| `CACHE.REDIS_ADDRESS`     | text   | —       | Server address(es). Separate multiple with commas for cluster mode. |
| `CACHE.REDIS_PREFIX`      | text   | —       | Key prefix to namespace keys when multiple apps share one instance. |
| `CACHE.REDIS_PASSWORD`    | text   | —       | Password for Redis authentication. Leave empty if not required. |
| `CACHE.REDIS_DB`          | number | `0`     | Database index (0–15). Applies to single-node only. |
| `CACHE.REDIS_DIAL_TIMEOUT`| text   | `5s`    | Timeout for establishing a connection. |
| `CACHE.REDIS_READ_TIMEOUT`| text   | `3s`    | Timeout for socket reads. |
| `CACHE.REDIS_WRITE_TIMEOUT`| text  | `3s`    | Timeout for socket writes. |

### config.yml

```yaml
Cache:
  RedisAddress: "localhost:6379"
  RedisPrefix: "myapp:"
  RedisPassword: "secret"
  RedisDb: 0
  RedisDialTimeout: "5s"
  RedisReadTimeout: "3s"
  RedisWriteTimeout: "3s"
```

### Cluster mode

Provide multiple addresses separated by commas to automatically use a `ClusterClient`:

```yaml
Cache:
  RedisAddress: "redis-1:6379,redis-2:6379,redis-3:6379"
  RedisPassword: "secret"
```

> **Note:** `CACHE.REDIS_DB` is ignored in cluster mode (clusters always use DB 0).

---

## Key-Value Cache

### Set

```go
// Store any value (serialized as JSON by default)
redis.Driver.Set("user:1", user)

// With TTL
redis.Driver.Set("session:abc", sessionData, 30*time.Minute)

// With bucket (key is stored as "bucket.key")
redis.Driver.Set("token", tok, kv.Bucket("auth"), 1*time.Hour)
```

### SetRaw

```go
redis.Driver.SetRaw("raw:key", []byte("binary data"))
redis.Driver.SetRaw("raw:key", []byte("binary data"), 10*time.Minute)
```

### SetNX — set if not exists

Atomically sets the key only if it does not already exist. Returns `true` if the key was written, `false` if it already existed.

```go
set, err := redis.Driver.SetNX("lock:payment:42", "locked", 30*time.Second)
if !set {
    // someone else holds the lock
}
```

### Get

```go
var user User
found := redis.Driver.Get("user:1", &user)
if !found {
    // cache miss
}
```

### GetRaw

```go
data, found := redis.Driver.GetRaw("raw:key")
```

### GetWithExpiration

Returns the value and its absolute expiration time. A zero `time.Time` means the key has no expiry.

```go
var user User
expireAt, found := redis.Driver.GetWithExpiration("user:1", &user)
if found && !expireAt.IsZero() {
    fmt.Println("expires at", expireAt)
}
```

### GetRawWithExpiration

```go
data, expireAt, found := redis.Driver.GetRawWithExpiration("raw:key")
```

### Exists

Check whether a key exists without fetching the value:

```go
if redis.Driver.Exists("session:abc") {
    // session is active
}
```

### Delete

```go
err := redis.Driver.Delete("user:1")
```

### Expire

Reset the expiration time of an existing key:

```go
err := redis.Driver.Expire("session:abc", time.Now().Add(10*time.Minute))
```

### Keys — pattern search

Scan for keys matching a glob pattern (uses Redis `SCAN`, never blocks):

```go
// All keys
keys, err := redis.Driver.Keys("*")

// All keys in a namespace
keys, err := redis.Driver.Keys("session:*")
```

> **Note:** In cluster mode only keys on the connected node are returned.

### Flush

Delete all keys in the current database:

```go
err := redis.Driver.Flush()
```

> **Note:** In cluster mode `Flush` only affects the connected node.

### ItemCount

Returns the total number of keys in the database (may include expired but not yet evicted keys):

```go
count := redis.Driver.ItemCount()
```

> **Note:** In cluster mode this reflects the connected node's count only.

---

## Increment / Decrement

Atomically increment or decrement an integer key by `n`:

```go
newVal, err := redis.Driver.Increment("page:views", 1)
newVal, err := redis.Driver.Decrement("credits:user:42", 5)
```

Supported `n` types: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`.

---

## Pub/Sub

### Publish

```go
// Publish any value (serialized as JSON)
err := redis.Driver.Publish("events.order", order)

// Publish raw bytes
err := redis.Driver.PublishBytes("events.order", []byte(`{"id":1}`))
```

### Subscribe

```go
redis.Driver.Subscribe("events.order", func(topic string, message []byte, driver pubsub.Interface) {
    var order Order
    driver.Unmarshal(message, &order)
    fmt.Println("received order", order.ID)
})
```

Multiple handlers can be registered for the same topic — all are invoked concurrently per message.

Subscribing to the same topic multiple times is safe. Only one goroutine is started per topic; additional handlers are appended to the existing subscription.

### UnsubscribeAll

Remove all handlers for a topic and stop its goroutine:

```go
redis.Driver.UnsubscribeAll("events.order")
```

### Reconnection

If the connection drops, the subscriber goroutine automatically resubscribes after a 2-second back-off. No manual intervention is required.

---

## Serializer

The default serializer is JSON. Override it with any type implementing `serializer.Interface`:

```go
import "github.com/getevo/evo/v2/lib/serializer"

redis.Driver.SetSerializer(serializer.MsgPack)
```

---

## Accessing the underlying client

The raw `redis.UniversalClient` is exposed for advanced use:

```go
import redisconn "github.com/getevo/evo/v2/lib/connectors/redis"

ctx := context.Background()

// Pipeline
pipe := redisconn.Client.Pipeline()
pipe.Set(ctx, "k1", "v1", 0)
pipe.Set(ctx, "k2", "v2", 0)
_, _ = pipe.Exec(ctx)

// Hash operations
redisconn.Client.HSet(ctx, "user:1", "name", "Alice", "age", 30)
name, _ := redisconn.Client.HGet(ctx, "user:1", "name").Result()

// Sorted set
redisconn.Client.ZAdd(ctx, "leaderboard", &redis.Z{Score: 9500, Member: "alice"})
```

---

## Cluster mode caveats

| Operation    | Behaviour in cluster mode |
|--------------|---------------------------|
| `Flush`      | Flushes only the connected node, not the whole cluster |
| `ItemCount`  | Returns the key count for the connected node only |
| `Keys`       | Scans only the connected node |
| `REDIS_DB`   | Ignored; clusters always use DB 0 |

---

## See Also

- [Configuration](configuration.md)
- [Database](database.md)
