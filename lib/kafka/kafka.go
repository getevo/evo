package kafka

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/segmentio/kafka-go"
	"net"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

type Message kafka.Message
type Partition kafka.Partition

type Client struct {
	_Brokers string
	Brokers  []string
}

func NewClient(servers string) *Client {
	return &Client{
		_Brokers: servers,
		Brokers:  strings.Split(servers, ","),
	}
}

func (client *Client) NewProducer(topic string, configs ...*ProducerConfig) *Producer {
	return NewProducer(client._Brokers, topic, configs...)
}

func (client *Client) NewConsumer(topic string, configs ...*ConsumerConfig) *Consumer {
	return NewConsumer(client._Brokers, topic, configs...)
}

func (client *Client) Topics() ([]string, error) {
	var topics []string
	if len(client.Brokers) == 0 {
		return topics, fmt.Errorf("invalid broker")
	}
	conn, err := kafka.Dial("tcp", client.Brokers[0])
	if err != nil {
		return topics, err
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return topics, err
	}

	m := map[string]struct{}{}
	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	for k := range m {
		topics = append(topics, k)
	}
	return topics, nil
}

func (client *Client) Partitions() ([]Partition, error) {
	var partitions []Partition
	if len(client.Brokers) == 0 {
		return partitions, fmt.Errorf("invalid broker")
	}
	conn, err := kafka.Dial("tcp", client.Brokers[0])
	if err != nil {
		return partitions, err
	}
	defer conn.Close()

	load, err := conn.ReadPartitions()
	if err != nil {
		return partitions, err
	}
	for _, item := range load {
		partitions = append(partitions, Partition(item))
	}
	return partitions, nil
}

func (client *Client) CreateTopic(topic string, replica, partitions int) error {
	if len(client.Brokers) == 0 {
		return fmt.Errorf("invalid broker")
	}
	conn, err := kafka.Dial("tcp", client.Brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replica,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return err
	}
	return nil
}

func applyConfig(src interface{}, dst interface{}) {

	var srcDesc = structs.New(src)
	var srcRef = reflect.ValueOf(src).Elem()
	var destRef = reflect.ValueOf(dst).Elem()

	for i := 0; i < srcRef.NumField(); i++ {
		rf := srcRef.Field(i)
		rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		if !rf.IsNil() {
			var name = srcDesc.Fields()[i].Name()
			dstField := destRef.FieldByName(name[1:])
			if dstField.IsValid() && dstField.CanSet() {
				dstField.Set(rf.Elem())
			}
		}
	}

}
