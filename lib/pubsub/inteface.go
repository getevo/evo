package pubsub

import "time"

type Message struct {
	Time    time.Time
	Message []byte
}

type Interface interface {
	Name() string
	Register() error
	Subscribe(topic string, onMessage func(topic string, message []byte, driver Interface), params ...interface{})
	Publish(topic string, message []byte, params ...interface{}) error

	// SetMarshaller set interface{} to []byte marshalling function
	SetMarshaller(func(input interface{}) ([]byte, error))

	// SetUnMarshaller set []byte to interface{} unmarshalling function
	SetUnMarshaller(func(bytes []byte, out interface{}) error)
}
