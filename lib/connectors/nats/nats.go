package nats

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/settings"
	"time"

	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/memo/kv"
	"github.com/getevo/evo/v2/lib/pubsub"
	"github.com/getevo/evo/v2/lib/serializer"
	"github.com/nats-io/nats.go"
)

var Driver = NATS{}

var listeners = map[string][]func(topic string, message []byte, driver pubsub.Interface){}
var _serializer = serializer.JSON
var prefix = ""
var nc *nats.Conn
var js nats.JetStreamContext

type NATS struct{}

func (d NATS) Subscribe(topic string, onMessage func(topic string, message []byte, driver pubsub.Interface), params ...any) {
	p := Parse(params)
	if !p.IgnorePrefix {
		topic = prefix + topic
	}

	if _, ok := listeners[topic]; !ok {
		listeners[topic] = []func(topic string, message []byte, driver pubsub.Interface){onMessage}
		nc.Subscribe(topic, func(msg *nats.Msg) {
			for _, callback := range listeners[topic] {
				go callback(topic, msg.Data, d)
			}
		})
	} else {
		listeners[topic] = append(listeners[topic], onMessage)
	}

}
func (d NATS) Publish(topic string, data any, params ...any) error {

	b, err := _serializer.Marshal(data)
	if err != nil {
		return err
	}
	return d.PublishBytes(topic, b, params...)
}

func (d NATS) PublishBytes(topic string, b []byte, params ...any) error {
	p := Parse(params)
	if !p.IgnorePrefix {
		topic = prefix + topic
	}
	return nc.Publish(topic, b)
}

func (NATS) Register() error {
	if nc != nil {
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
			Description: "List of comma separated NATS servers",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},

		settings.Setting{
			Domain:      "NATS",
			Name:        "MAX_RECONNECT",
			Title:       "Max Reconnects",
			Description: "MaxReconnect attempts",
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
			Description: "Wait time before reconnect",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},

		settings.Setting{
			Domain:      "NATS",
			Name:        "RANDOMIZE",
			Title:       "Reconnect Wait",
			Description: "Randomization of the server pool",
			Type:        "bool",
			Value:       "true",
			ReadOnly:    false,
			Visible:     true,
		},

		settings.Setting{
			Domain:      "NATS",
			Name:        "USERNAME",
			Title:       "Username",
			Description: "Username in case of authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},

		settings.Setting{
			Domain:      "NATS",
			Name:        "PASSWORD",
			Title:       "Password",
			Description: "Password in case of authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},

		settings.Setting{
			Domain:      "NATS",
			Name:        "TOKEN",
			Title:       "Token",
			Description: "Token in case of authentication",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
	)
	var err error
	var options []nats.Option
	if !settings.Get("NATS.RANDOMIZE").Bool() {
		options = append(options, nats.DontRandomize())
	}

	if settings.Get("NATS.USERNAME").String() != "" {
		options = append(options, nats.UserInfo(settings.Get("NATS.USERNAME").String(), settings.Get("NATS.PASSWORD").String()))
	}

	if settings.Get("NATS.TOKEN").String() != "" {
		options = append(options, nats.Token(settings.Get("NATS.TOKEN").String()))
	}

	if wait, err := settings.Get("NATS.RECONNECT_WAIT").Duration(); wait > 0 && err == nil {
		options = append(options, nats.ReconnectWait(wait))
	}

	if reconnects := settings.Get("NATS.MAX_RECONNECT").Int(); reconnects > -1 {
		options = append(options, nats.MaxReconnects(reconnects))
	}

	nc, err = nats.Connect(settings.Get("NATS.SERVER").String(), options...)
	if err != nil {
		return err
	}
	js, err = nc.JetStream()

	return err
}

func (NATS) Name() string {
	return "nats"
}

// SetSerializer set data serialization method
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

func (d NATS) Set(key string, value any, params ...any) error {
	var p = kv.Parse(params)
	b, err := _serializer.Marshal(value)
	if err != nil {
		return err
	}
	var kv nats.KeyValue
	kv, err = d.GetKV(p.Bucket)
	if err != nil {
		return err
	}
	_, err = kv.Put(key, b)
	return err
}

func (d NATS) SetRaw(key string, value []byte, params ...any) error {
	var p = kv.Parse(params)
	kv, err := d.GetKV(p.Bucket)
	if err != nil {
		return err
	}
	_, err = kv.Put(key, value)
	return err
}

func (d NATS) Replace(key string, value any, params ...any) bool {
	var p = kv.Parse(params)
	b, err := _serializer.Marshal(value)
	if err != nil {
		return false
	}
	var kv nats.KeyValue
	kv, err = d.GetKV(p.Bucket)
	if err != nil {
		return false
	}
	_, err = kv.Put(key, b)
	return err == nil
}

func (d NATS) Get(key string, out any, params ...any) bool {
	var p = kv.Parse(params)
	kv, err := d.GetKV(p.Bucket)
	if err != nil {
		return false
	}
	entry, err := kv.Get(key)
	if err != nil {
		return false
	}
	return _serializer.Unmarshal(entry.Value(), out) == nil
}

func (d NATS) GetRaw(key string, params ...any) ([]byte, bool) {
	var p = kv.Parse(params)
	kv, err := d.GetKV(p.Bucket)
	if err != nil {
		return []byte{}, false
	}
	entry, err := kv.Get(key)
	if err != nil {
		return []byte{}, false
	}
	return entry.Value(), true
}

func (d NATS) GetWithExpiration(key string, out any, params ...any) (time.Time, bool) {
	log.Error(d.Name() + " does not support expiration")
	return time.Time{}, false
}

func (d NATS) GetRawWithExpiration(key string, params ...any) ([]byte, time.Time, bool) {
	log.Error(d.Name() + " does not support expiration")
	return []byte{}, time.Time{}, false
}

func (d NATS) Increment(key string, n any, params ...any) (int64, error) {
	var err = fmt.Errorf(d.Name() + " does not support increment")
	log.Error(err)
	return 0, err
}

func (d NATS) Decrement(key string, n any, params ...any) (int64, error) {
	var err = fmt.Errorf(d.Name() + " does not support decrement")
	log.Error(err)
	return 0, err
}

func (d NATS) Delete(key string, params ...any) error {
	var p = kv.Parse(params)
	kv, err := d.GetKV(p.Bucket)
	if err != nil {
		return err
	}
	return kv.Delete(key)
}

func (d NATS) Expire(key string, t time.Time, params ...any) error {
	var err = fmt.Errorf(d.Name() + " does not support expiration")
	log.Error(err)
	return err
}

func (d NATS) ItemCount() int64 {
	var err = fmt.Errorf(d.Name() + " does not support Item Count")
	log.Error(err)
	return 0
}

func (d NATS) Flush() error {
	var err = fmt.Errorf(d.Name() + " does not support Item Count")
	log.Error(err)
	return err
}

var list = map[string]nats.KeyValue{}

func (d NATS) GetKV(bucket string) (nats.KeyValue, error) {
	if v, ok := list[bucket]; ok {
		return v, nil
	}
	var kv nats.KeyValue
	var err error
	if stream, _ := js.StreamInfo(bucket); stream == nil {
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: bucket,
		})
		if err != nil {
			return nil, err
		}
	} else {
		kv, err = js.KeyValue(bucket)
		if err != nil {
			return nil, err
		}
	}
	list[bucket] = kv
	return kv, nil
}
