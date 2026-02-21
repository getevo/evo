package nats

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/memo/kv"
	"github.com/getevo/evo/v2/lib/pubsub"
	"github.com/getevo/evo/v2/lib/serializer"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/getevo/evo/v2/lib/shutdown"
	"github.com/nats-io/nats.go"
)

var Driver = NATS{}

// Connection is the underlying NATS connection. Available after Register succeeds.
var Connection *nats.Conn

// JS is the JetStream context. Available after Register succeeds.
var JS nats.JetStreamContext

var (
	listeners    = map[string][]func(topic string, message []byte, driver pubsub.Interface){}
	listenersMu  sync.RWMutex
	_serializer  = serializer.JSON
	prefix       = ""
	list         = map[string]nats.KeyValue{}
	listMu       sync.RWMutex
	defaultBucket = "default"
)

var errNotReady = errors.New("nats: not connected — call Register() first")

type NATS struct{}

// Subscribe registers onMessage for the given topic.
// Pass nats.Queue("group") to use a queue-subscribe group (load-balanced).
// Pass nats.IgnorePrefix to skip the configured topic prefix.
func (d NATS) Subscribe(topic string, onMessage func(topic string, message []byte, driver pubsub.Interface), params ...any) {
	if Connection == nil {
		log.Error(errNotReady)
		return
	}

	p := Parse(params)
	if !p.IgnorePrefix {
		topic = prefix + topic
	}

	listenersMu.Lock()
	if _, ok := listeners[topic]; !ok {
		listeners[topic] = []func(topic string, message []byte, driver pubsub.Interface){onMessage}
		listenersMu.Unlock()

		handler := func(msg *nats.Msg) {
			listenersMu.RLock()
			cbs := make([]func(string, []byte, pubsub.Interface), len(listeners[msg.Subject]))
			copy(cbs, listeners[msg.Subject])
			listenersMu.RUnlock()
			for _, cb := range cbs {
				go cb(msg.Subject, msg.Data, d)
			}
		}

		var err error
		if p.QueueGroup != "" {
			_, err = Connection.QueueSubscribe(topic, p.QueueGroup, handler)
		} else {
			_, err = Connection.Subscribe(topic, handler)
		}
		if err != nil {
			log.Error("nats: failed to subscribe", "topic", topic, "error", err)
			// clean up the entry so a retry can succeed
			listenersMu.Lock()
			delete(listeners, topic)
			listenersMu.Unlock()
		}
	} else {
		listeners[topic] = append(listeners[topic], onMessage)
		listenersMu.Unlock()
	}
}

// Publish serializes data and publishes it to topic.
// Pass nats.WithJetStream to use JetStream publish (at-least-once delivery).
// Pass nats.IgnorePrefix to skip the configured topic prefix.
func (d NATS) Publish(topic string, data any, params ...any) error {
	b, err := _serializer.Marshal(data)
	if err != nil {
		return err
	}
	return d.PublishBytes(topic, b, params...)
}

// PublishBytes publishes raw bytes to topic.
// Pass nats.WithJetStream to use JetStream publish (at-least-once delivery).
// Pass nats.IgnorePrefix to skip the configured topic prefix.
func (d NATS) PublishBytes(topic string, b []byte, params ...any) error {
	if Connection == nil {
		return errNotReady
	}
	p := Parse(params)
	if !p.IgnorePrefix {
		topic = prefix + topic
	}
	if p.UseJetStream {
		if JS == nil {
			return errors.New("nats: JetStream not available")
		}
		_, err := JS.Publish(topic, b)
		return err
	}
	return Connection.Publish(topic, b)
}

func (NATS) Register() error {
	if Connection != nil && JS != nil {
		return nil
	}

	_ = settings.Register(
		settings.SettingDomain{
			Title:       "NATS",
			Domain:      "NATS",
			Description: "NATS configurations",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "SERVER",
			Title:       "Servers",
			Description: "Comma-separated list of NATS server URLs",
			Type:        "text",
			Value:       nats.DefaultURL,
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "MAX_RECONNECT",
			Title:       "Max Reconnects",
			Description: "Maximum reconnect attempts (-1 = unlimited)",
			Type:        "number",
			Value:       "5",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "RECONNECT_WAIT",
			Title:       "Reconnect Wait",
			Value:       "2s",
			Description: "Wait duration between reconnect attempts",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "CONNECT_TIMEOUT",
			Title:       "Connect Timeout",
			Value:       "5s",
			Description: "Timeout for initial connection attempt",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "DRAIN_TIMEOUT",
			Title:       "Drain Timeout",
			Value:       "30s",
			Description: "Maximum time to wait for in-flight messages to drain on shutdown",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "RANDOMIZE",
			Title:       "Randomize",
			Description: "Randomize the server pool on connect",
			Type:        "bool",
			Value:       "true",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "USERNAME",
			Title:       "Username",
			Description: "Username for password authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "PASSWORD",
			Title:       "Password",
			Description: "Password for password authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "TOKEN",
			Title:       "Token",
			Description: "Token for token authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "CREDENTIALS",
			Title:       "Credentials File",
			Description: "Path to a .creds file for JWT+NKey authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "NKEY_FILE",
			Title:       "NKey Seed File",
			Description: "Path to an NKey seed file for NKey authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "TLS_CERT",
			Title:       "TLS Certificate",
			Description: "Path to client TLS certificate file (PEM)",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "TLS_KEY",
			Title:       "TLS Key",
			Description: "Path to client TLS key file (PEM)",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "TLS_CA",
			Title:       "TLS CA Certificate",
			Description: "Path to CA certificate file for server verification (PEM)",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "NATS",
			Name:        "DEFAULT_BUCKET",
			Title:       "Default KV Bucket",
			Description: "Default JetStream KV bucket name used when none is specified",
			Type:        "text",
			Value:       "default",
			ReadOnly:    false,
			Visible:     true,
		},
	)

	if b := settings.Get("NATS.DEFAULT_BUCKET").String(); b != "" {
		defaultBucket = b
	}

	var err error
	if Connection == nil {
		var options []nats.Option

		// Connection event handlers for observability
		options = append(options, nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				log.Error("nats: disconnected", "error", err)
			} else {
				log.Warning("nats: disconnected")
			}
		}))
		options = append(options, nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("nats: reconnected", "url", nc.ConnectedUrl())
		}))
		options = append(options, nats.ClosedHandler(func(_ *nats.Conn) {
			log.Warning("nats: connection closed")
		}))
		options = append(options, nats.ErrorHandler(func(_ *nats.Conn, sub *nats.Subscription, err error) {
			subject := ""
			if sub != nil {
				subject = sub.Subject
			}
			log.Error("nats: async error", "subject", subject, "error", err)
		}))

		// Server pool
		if !settings.Get("NATS.RANDOMIZE").Bool() {
			options = append(options, nats.DontRandomize())
		}

		// Timeouts
		if timeout, err := settings.Get("NATS.CONNECT_TIMEOUT").Duration(); timeout > 0 && err == nil {
			options = append(options, nats.Timeout(timeout))
		}
		if wait, err := settings.Get("NATS.RECONNECT_WAIT").Duration(); wait > 0 && err == nil {
			options = append(options, nats.ReconnectWait(wait))
		}
		if reconnects := settings.Get("NATS.MAX_RECONNECT").Int(); reconnects > -1 {
			options = append(options, nats.MaxReconnects(reconnects))
		}
		if drainTimeout, err := settings.Get("NATS.DRAIN_TIMEOUT").Duration(); drainTimeout > 0 && err == nil {
			options = append(options, nats.DrainTimeout(drainTimeout))
		}

		// Authentication — credentials file takes precedence, then nkey, then token, then user/pass
		if creds := settings.Get("NATS.CREDENTIALS").String(); creds != "" {
			options = append(options, nats.UserCredentials(creds))
		} else if nkeyFile := settings.Get("NATS.NKEY_FILE").String(); nkeyFile != "" {
			opt, err := nats.NkeyOptionFromSeed(nkeyFile)
			if err != nil {
				return fmt.Errorf("nats: invalid nkey file: %w", err)
			}
			options = append(options, opt)
		} else if token := settings.Get("NATS.TOKEN").String(); token != "" {
			options = append(options, nats.Token(token))
		} else if user := settings.Get("NATS.USERNAME").String(); user != "" {
			options = append(options, nats.UserInfo(user, settings.Get("NATS.PASSWORD").String()))
		}

		// TLS
		tlsCert := settings.Get("NATS.TLS_CERT").String()
		tlsKey := settings.Get("NATS.TLS_KEY").String()
		if tlsCert != "" && tlsKey != "" {
			options = append(options, nats.ClientCert(tlsCert, tlsKey))
		}
		if ca := settings.Get("NATS.TLS_CA").String(); ca != "" {
			options = append(options, nats.RootCAs(ca))
		}

		Connection, err = nats.Connect(settings.Get("NATS.SERVER").String(), options...)
		if err != nil {
			return err
		}
		log.Info("nats: connected", "url", Connection.ConnectedUrl())
	}

	JS, err = Connection.JetStream()
	if err != nil {
		return err
	}

	// Graceful shutdown: drain in-flight messages before the process exits.
	shutdown.Register(func() {
		if Connection == nil || Connection.IsClosed() {
			return
		}
		log.Info("nats: draining connection")
		if err := Connection.Drain(); err != nil {
			log.Error("nats: drain failed", "error", err)
		}
	})

	return nil
}

func (NATS) Name() string {
	return "nats"
}

// SetSerializer sets the data serialization method.
func (NATS) SetSerializer(v serializer.Interface) {
	_serializer = v
}

func (NATS) SetPrefix(s string) {
	prefix = s
}

func (NATS) Serializer() serializer.Interface {
	return _serializer
}

func (NATS) Marshal(data any) ([]byte, error) {
	return _serializer.Marshal(data)
}

func (NATS) Unmarshal(data []byte, v any) error {
	return _serializer.Unmarshal(data, v)
}

// ── KV store ─────────────────────────────────────────────────────────────────

func (d NATS) Set(key string, value any, params ...any) error {
	p := kv.Parse(params)
	b, err := _serializer.Marshal(value)
	if err != nil {
		return err
	}
	store, err := d.GetKV(p.Bucket)
	if err != nil {
		return err
	}
	_, err = store.Put(key, b)
	return err
}

func (d NATS) SetRaw(key string, value []byte, params ...any) error {
	p := kv.Parse(params)
	store, err := d.GetKV(p.Bucket)
	if err != nil {
		return err
	}
	_, err = store.Put(key, value)
	return err
}

// Replace sets key only if it already exists in the bucket.
func (d NATS) Replace(key string, value any, params ...any) bool {
	p := kv.Parse(params)
	store, err := d.GetKV(p.Bucket)
	if err != nil {
		return false
	}
	entry, err := store.Get(key)
	if err != nil {
		return false // key does not exist
	}
	b, err := _serializer.Marshal(value)
	if err != nil {
		return false
	}
	// Use Update (CAS on last revision) to honour the "only if exists" contract.
	_, err = store.Update(key, b, entry.Revision())
	return err == nil
}

func (d NATS) Get(key string, out any, params ...any) bool {
	p := kv.Parse(params)
	store, err := d.GetKV(p.Bucket)
	if err != nil {
		return false
	}
	entry, err := store.Get(key)
	if err != nil {
		return false
	}
	return _serializer.Unmarshal(entry.Value(), out) == nil
}

func (d NATS) GetRaw(key string, params ...any) ([]byte, bool) {
	p := kv.Parse(params)
	store, err := d.GetKV(p.Bucket)
	if err != nil {
		return nil, false
	}
	entry, err := store.Get(key)
	if err != nil {
		return nil, false
	}
	return entry.Value(), true
}

func (d NATS) GetWithExpiration(key string, out any, params ...any) (time.Time, bool) {
	log.Error(d.Name() + " does not support per-key expiration")
	return time.Time{}, false
}

func (d NATS) GetRawWithExpiration(key string, params ...any) ([]byte, time.Time, bool) {
	log.Error(d.Name() + " does not support per-key expiration")
	return nil, time.Time{}, false
}

func (d NATS) Increment(key string, n any, params ...any) (int64, error) {
	err := fmt.Errorf(d.Name() + " does not support increment")
	log.Error(err)
	return 0, err
}

func (d NATS) Decrement(key string, n any, params ...any) (int64, error) {
	err := fmt.Errorf(d.Name() + " does not support decrement")
	log.Error(err)
	return 0, err
}

func (d NATS) Delete(key string, params ...any) error {
	p := kv.Parse(params)
	store, err := d.GetKV(p.Bucket)
	if err != nil {
		return err
	}
	return store.Delete(key)
}

func (d NATS) Expire(key string, t time.Time, params ...any) error {
	err := fmt.Errorf(d.Name() + " does not support per-key expiration")
	log.Error(err)
	return err
}

func (d NATS) ItemCount() int64 {
	log.Error(d.Name() + " does not support ItemCount")
	return 0
}

func (d NATS) Flush() error {
	err := fmt.Errorf(d.Name() + " does not support Flush")
	log.Error(err)
	return err
}

// GetKV returns the JetStream KeyValue handle for bucket, creating it if it
// does not yet exist. Use CreateBucket for full control over bucket config.
func (d NATS) GetKV(bucket string) (nats.KeyValue, error) {
	if JS == nil {
		return nil, errNotReady
	}
	if bucket == "" {
		bucket = defaultBucket
	}

	listMu.RLock()
	v, ok := list[bucket]
	listMu.RUnlock()
	if ok {
		return v, nil
	}

	listMu.Lock()
	defer listMu.Unlock()

	// Double-check after acquiring the write lock.
	if v, ok := list[bucket]; ok {
		return v, nil
	}

	store, err := JS.KeyValue(bucket)
	if errors.Is(err, nats.ErrBucketNotFound) {
		store, err = JS.CreateKeyValue(&nats.KeyValueConfig{Bucket: bucket})
	}
	if err != nil {
		return nil, err
	}

	list[bucket] = store
	return store, nil
}

// CreateBucket creates (or updates) a JetStream KV bucket with full config.
// Call this before any Set/Get if you need non-default settings such as TTL,
// replication factor, storage type, or key history.
func (d NATS) CreateBucket(cfg *nats.KeyValueConfig) error {
	if JS == nil {
		return errNotReady
	}
	if cfg.Bucket == "" {
		return errors.New("nats: bucket name must not be empty")
	}

	listMu.Lock()
	defer listMu.Unlock()

	store, err := JS.CreateKeyValue(cfg)
	if err != nil {
		return err
	}
	list[cfg.Bucket] = store
	return nil
}
