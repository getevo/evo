# NATS Connector

The NATS connector provides two capabilities backed by a single NATS connection:

- **Pub/Sub** — implements `pubsub.Interface` for message broadcasting
- **Key/Value store** — implements `memo/kv.Interface` using JetStream KV

---

## Installation

```go
import "github.com/getevo/evo/v2/lib/connectors/nats"
```

---

## Configuration

Settings are registered automatically under the `NATS` domain when `Register()` is called.

| Key                    | Type   | Default                    | Description                                                    |
|------------------------|--------|----------------------------|----------------------------------------------------------------|
| `NATS.SERVER`          | text   | `nats://127.0.0.1:4222`    | Comma-separated list of NATS server URLs                       |
| `NATS.MAX_RECONNECT`   | number | `5`                        | Maximum reconnect attempts (`-1` = unlimited)                  |
| `NATS.RECONNECT_WAIT`  | text   | `2s`                       | Wait duration between reconnect attempts                       |
| `NATS.CONNECT_TIMEOUT` | text   | `5s`                       | Timeout for the initial connection attempt                     |
| `NATS.DRAIN_TIMEOUT`   | text   | `30s`                      | Max time to wait for in-flight messages to drain on shutdown   |
| `NATS.RANDOMIZE`       | bool   | `true`                     | Randomize the server pool on connect                           |
| `NATS.DEFAULT_BUCKET`  | text   | `default`                  | KV bucket used when no `Bucket()` param is passed              |

### Authentication

Authentication methods are mutually exclusive. The first non-empty setting wins, in this order:

| Key                  | Type   | Description                                          |
|----------------------|--------|------------------------------------------------------|
| `NATS.CREDENTIALS`   | text   | Path to a `.creds` file (JWT + NKey)                 |
| `NATS.NKEY_FILE`     | text   | Path to an NKey seed file                            |
| `NATS.TOKEN`         | text   | Static token                                         |
| `NATS.USERNAME`      | text   | Username (used together with `NATS.PASSWORD`)        |
| `NATS.PASSWORD`      | text   | Password                                             |

### TLS

| Key           | Type | Description                              |
|---------------|------|------------------------------------------|
| `NATS.TLS_CERT` | text | Path to client TLS certificate (PEM)   |
| `NATS.TLS_KEY`  | text | Path to client TLS key (PEM)           |
| `NATS.TLS_CA`   | text | Path to CA certificate for server verification (PEM) |

Example settings file:

```
NATS.SERVER=nats://nats1:4222,nats://nats2:4222
NATS.MAX_RECONNECT=-1
NATS.RECONNECT_WAIT=5s
NATS.CREDENTIALS=/etc/nats/app.creds
NATS.DEFAULT_BUCKET=myapp
```

---

## Pub/Sub

### Register as the default pub/sub driver

```go
import (
    "github.com/getevo/evo/v2/lib/pubsub"
    natspkg "github.com/getevo/evo/v2/lib/connectors/nats"
)

pubsub.SetDefaultDriver(natspkg.Driver)
```

### Subscribe

```go
pubsub.Subscribe("orders.created", func(topic string, message []byte, driver pubsub.Interface) {
    var order Order
    driver.Unmarshal(message, &order)
    // handle order...
})
```

### Queue subscribe (load-balanced)

Messages are delivered to exactly one member of the queue group.

```go
natspkg.Driver.Subscribe("orders.created", handler, natspkg.Queue("workers"))
```

### Publish (core NATS — fire and forget)

```go
pubsub.Publish("orders.created", order)
```

### Publish via JetStream (at-least-once delivery)

```go
natspkg.Driver.Publish("orders.created", order, natspkg.WithJetStream)
```

The server acknowledges receipt. Returns an error if the message is not persisted.
The subject must match a JetStream stream configured on the server.

### Publish raw bytes

```go
pubsub.PublishBytes("orders.created", []byte(`{"id":1}`))
```

### Topic prefix

All topics are automatically prefixed with the value set via `SetPrefix`. To bypass the prefix for a single call:

```go
natspkg.Driver.Subscribe("internal.topic", handler, natspkg.IgnorePrefix)
natspkg.Driver.Publish("internal.topic", payload, natspkg.IgnorePrefix)
```

### Change serializer

The default serializer is JSON. To switch:

```go
import "github.com/getevo/evo/v2/lib/serializer"
natspkg.Driver.SetSerializer(serializer.MsgPack)
```

---

## JetStream Key/Value Store

Buckets are created automatically on first use with default settings. Call
`CreateBucket` first if you need custom TTL, replication, storage type, or
key history.

### Set a value

```go
natspkg.Driver.Set("user:1", userStruct, natspkg.Bucket("my-bucket"))
```

### Get a value

```go
var user User
ok := natspkg.Driver.Get("user:1", &user, natspkg.Bucket("my-bucket"))
```

### Get raw bytes

```go
raw, ok := natspkg.Driver.GetRaw("user:1", natspkg.Bucket("my-bucket"))
```

### Delete a key

```go
err := natspkg.Driver.Delete("user:1", natspkg.Bucket("my-bucket"))
```

### Replace (update only if key exists)

`Replace` returns `false` and makes no change if the key is not already present.

```go
ok := natspkg.Driver.Replace("user:1", updatedUser, natspkg.Bucket("my-bucket"))
```

### Default bucket

When no `Bucket()` param is passed, the bucket configured via `NATS.DEFAULT_BUCKET`
(default: `"default"`) is used:

```go
natspkg.Driver.Set("session:abc", sessionData) // uses "default" bucket
```

### Create a bucket with custom settings

```go
import "github.com/nats-io/nats.go"

err := natspkg.Driver.CreateBucket(&nats.KeyValueConfig{
    Bucket:   "sessions",
    TTL:      24 * time.Hour,  // stream-level max-age
    Replicas: 3,
    History:  5,               // keep last 5 revisions per key
})
```

After `CreateBucket` the bucket is cached internally; subsequent `Set`/`Get` calls
use the pre-created configuration.

---

## Graceful shutdown

The NATS connector automatically registers a shutdown hook via `evo.OnShutdown`.
When `evo.Shutdown()` is called (or SIGTERM / SIGINT is received), the connector
calls `nc.Drain()` before the process exits, ensuring that:

- All pending publish operations are flushed to the server.
- All active subscriptions finish processing in-flight messages.

The maximum drain duration is controlled by `NATS.DRAIN_TIMEOUT` (default `30s`).

---

## Raw access

The underlying connection and JetStream context are exported for advanced use cases:

```go
// Use the raw NATS connection (request/reply, object store, etc.)
msg, err := natspkg.Connection.Request("service.ping", nil, time.Second)

// Use the raw JetStream context
info, err := natspkg.JS.StreamInfo("MY_STREAM")
```

---

## Limitations

The following `memo.Interface` methods are not supported by JetStream KV:

| Method                 | Reason                                       |
|------------------------|----------------------------------------------|
| `Expire`               | JetStream KV has no per-key TTL              |
| `GetWithExpiration`    | JetStream KV has no per-key TTL              |
| `GetRawWithExpiration` | JetStream KV has no per-key TTL              |
| `Increment`            | Not supported by JetStream KV                |
| `Decrement`            | Not supported by JetStream KV                |
| `ItemCount`            | Not supported by this driver                 |
| `Flush`                | Not supported by this driver                 |

> **Tip:** Stream-level expiry (max age for all keys in a bucket) can be set via
> `CreateBucket` using the `TTL` field of `nats.KeyValueConfig`.
