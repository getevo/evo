package kafka

import (
	"github.com/getevo/evo/v2/lib/kafka"
	"github.com/getevo/evo/v2/lib/pubsub"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/kelindar/binary"
)

var Driver = driver{}
var marshaller func(input interface{}) ([]byte, error) = binary.Marshal
var unmarshaller func(bytes []byte, out interface{}) error = binary.Unmarshal
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
var Client *kafka.Client

func (driver) Register() error {
	if Client != nil {
		return nil
	}
	settings.Register(
		settings.SettingDomain{
			Title:       "Kafka",
			Domain:      "KAFKA",
			Description: "Apache Kafka configurations",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "BROKERS",
			Title:       "Kafka Brokers",
			Description: "List of comma separated kafka brokers",
			Type:        "text",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "COMPRESSION",
			Title:       "Compression Method",
			Description: "Any of (none,gzip,snappy,lz4,zstd)",
			Type:        "select",
			Params:      "[none,gzip,snappy,lz4,zstd]",
			Value:       "none",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "BALANCER",
			Title:       "Balancer",
			Description: "Balancer strategy (MurMur2,CRC32,Hash,LeastBytes,RoundRobin)",
			Type:        "select",
			Value:       "RoundRobin",
			Params:      "[MurMur2,CRC32,Hash,LeastBytes,RoundRobin]",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "REQUIRE_ACKS",
			Title:       "Require Acks",
			Description: "Number of acknowledges from partition replicas required before receiving.",
			Type:        "select",
			Params:      "[RequireNone,RequireOne,RequireAll]",
			Value:       "RequireNone",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "MaxAttempts",
			Title:       "Max Attempts",
			Description: "Maximum number of attempts to send a message",
			Value:       "10",
			Type:        "number",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "BatchBytes",
			Title:       "Batch Bytes",
			Description: "Limit the maximum size of a request in bytes before being sent to a partition.",
			Type:        "number",
			Value:       "1048576",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "BatchSize",
			Title:       "Batch Size",
			Description: "Limit on how many messages will be buffered before being sent to a partition",
			Type:        "number",
			Value:       "100",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "BatchTimeout",
			Title:       "Batch Timeout",
			Description: "Time limit on how often incomplete message batches will be flushed to kafka",
			Type:        "duration",
			Value:       "1s",
			ReadOnly:    false,
			Visible:     true,
		},

		settings.Setting{
			Domain:      "KAFKA",
			Name:        "ReadTimeout",
			Title:       "Read Timeout",
			Description: "Timeout for read operations performed by the Writer.",
			Type:        "duration",
			Value:       "10s",
			ReadOnly:    false,
			Visible:     true,
		},

		settings.Setting{
			Domain:      "KAFKA",
			Name:        "WriteTimeout",
			Title:       "Write Timeout",
			Description: "Timeout for write operations performed by the Writer.",
			Type:        "duration",
			Value:       "10s",
			ReadOnly:    false,
			Visible:     true,
		},
		settings.Setting{
			Domain:      "KAFKA",
			Name:        "ASYNC_WRITE",
			Title:       "Async Write",
			Description: "Cause non blocking but untraceable writes",
			Type:        "select",
			Params:      "[true,false]",
			Value:       "true",
			ReadOnly:    false,
			Visible:     true,
		},
	)

	Client = kafka.NewClient(settings.Get("KAFKA.BROKERS").String())

	return nil
}

func (driver) Name() string {
	return "kafka"
}

// SetMarshaller set interface{} to []byte marshalling function
func (driver) SetMarshaller(fn func(input interface{}) ([]byte, error)) {
	marshaller = fn
}

// SetUnMarshaller set []byte to interface{} unmarshalling function
func (driver) SetUnMarshaller(fn func(bytes []byte, out interface{}) error) {
	unmarshaller = fn
}
