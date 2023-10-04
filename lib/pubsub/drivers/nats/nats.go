package nats

import (
	"github.com/getevo/evo/v2/lib/kafka"
	"github.com/getevo/evo/v2/lib/pubsub"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/nats-io/nats.go"
)

var Driver = driver{}

var listeners = map[string][]func(topic string, message []byte, driver pubsub.Interface){}
var producers = map[string]*kafka.Producer{}

type driver struct{}

func (d driver) Subscribe(topic string, onMessage func(topic string, message []byte, driver pubsub.Interface), params ...interface{}) {
	if _, ok := listeners[topic]; !ok {
		listeners[topic] = []func(topic string, message []byte, driver pubsub.Interface){onMessage}
		var configs []*kafka.ConsumerConfig
		for idx, _ := range params {
			if v, ok := params[idx].(kafka.ConsumerConfig); ok {
				configs = append(configs, &v)
				continue
			}
			if v, ok := params[idx].(*kafka.ConsumerConfig); ok {
				configs = append(configs, v)
			}
		}
		Client.NewConsumer(topic, configs...).OnMessage(func(message kafka.Message) {
			for _, callback := range listeners[topic] {
				go callback(topic, message.Value, d)
			}
		})
	} else {
		listeners[topic] = append(listeners[topic], onMessage)
	}

}
func (d driver) Publish(topic string, message []byte, params ...interface{}) error {
	if _, ok := producers[topic]; !ok {
		var config = kafka.ProducerConfig{}
		switch settings.Get("KAFKA.COMPRESSION").String() {
		case "gzip":
			config.Compression(kafka.Gzip)
		case "snappy":
			config.Compression(kafka.Snappy)
		case "lz4":
			config.Compression(kafka.Lz4)
		case "zstd":
			config.Compression(kafka.Zstd)
		}
		switch settings.Get("KAFKA.BALANCER").String() {
		case "MurMur2":
			config.Balancer(kafka.MurMur2)
		case "CRC32":
			config.Balancer(kafka.CRC32)
		case "Hash":
			config.Balancer(kafka.Hash)
		case "LeastBytes":
			config.Balancer(kafka.LeastBytes)
		case "RoundRobin":
			config.Balancer(kafka.RoundRobin)
		}
		switch settings.Get("KAFKA.REQUIRE_ACKS").String() {
		case "RequireNone":
			config.Ack(kafka.RequireNone)
		case "RequireOne":
			config.Ack(kafka.RequireOne)
		case "RequireAll":
			config.Ack(kafka.RequireAll)
		}
		readTimeout, err := settings.Get("KAFKA.BatchTimeout").Duration()
		if err != nil {
			config.ReadTimeout(readTimeout)
		}
		writeTimeout, err := settings.Get("KAFKA.WriteTimeout").Duration()
		if err != nil {
			config.WriteTimeout(writeTimeout)
		}

		batchTimeout, err := settings.Get("KAFKA.BatchTimeout").Duration()
		if err != nil {
			config.BatchTimeout(batchTimeout)
		}
		config.MaxAttempts(settings.Get("KAFKA.MaxAttempts").Int())
		config.BatchSize(settings.Get("KAFKA.BatchSize").Int())
		config.BatchBytes(settings.Get("KAFKA.BatchBytes").Int64())
		config.Async(settings.Get("KAFKA.ASYNC_WRITE").Bool())
		producers[topic] = Client.NewProducer(topic, &config)
	}
	return producers[topic].Write(kafka.Message{
		Value: message,
	})
}

var prefix = ""
var Client *nats.Conn

func (driver) Register() error {
	if Client != nil {
		return nil
	}
	settings.Register(
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

	nats.CustomInboxPrefix()
	Client, err = nats.Connect(settings.Get("NATS.SERVER").String(), options...)

	return err
}

func (driver) Name() string {
	return "kafka"
}

// SetMarshaller set interface{} to []byte marshalling function
func (driver) SetMarshaller(fn func(input interface{}) ([]byte, error)) {

}

// SetUnMarshaller set []byte to interface{} unmarshalling function
func (driver) SetUnMarshaller(fn func(bytes []byte, out interface{}) error) {

}
