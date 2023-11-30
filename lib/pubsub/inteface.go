package pubsub

import (
	"github.com/getevo/evo/v2/lib/serializer"
	"time"
)

type Message struct {
	Time    time.Time
	Message []byte
}

type Interface interface {
	Name() string
	Register() error
	Subscribe(topic string, onMessage func(topic string, message []byte, driver Interface), params ...any)
	Publish(topic string, message any, params ...any) error
	PublishBytes(topic string, message []byte, params ...any) error

	// SetSerializer change serialization method
	SetSerializer(v serializer.Interface)

	Serializer() serializer.Interface

	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error

	SetPrefix(s string)
}
