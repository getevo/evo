package kafka_test

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/kafka"
	"github.com/getevo/evo/v2/lib/text"
	"testing"
	"time"
)

func TestKafka(t *testing.T) {
	var config = kafka.NewConsumerConfig()
	var brokers = "10.11.10.100:31092,10.11.10.101:31092,10.11.10.102:31092,10.11.10.103:31093"
	consumer := kafka.NewConsumer(brokers, "samsa_roundone", config)
	consumer.OnMessage(func(message kafka.Message) {
		fmt.Println(message.Topic, string(message.Key), string(message.Value))
	})
	fmt.Printf("%+v \n", consumer.Config)

	var cfg = kafka.NewProducerConfig().BatchSize(0).BatchTimeout(100 * time.Millisecond)
	producer := kafka.NewProducer(brokers, "samsa_roundone", cfg)

	fmt.Printf("%+v \n", *producer.Writer)
	c := 0
	for {
		c++
		producer.Write(kafka.Message{
			Value: []byte(fmt.Sprint(time.Now().Unix())),
		})
		if c == 10 {
			fmt.Printf("%+v \n", producer.Writer.Stats())
			break
		}
	}

	for {
		time.Sleep(10 * time.Second)
	}

}

func TestClient(t *testing.T) {
	var brokers = "10.11.10.100:31092,10.11.10.101:31092,10.11.10.102:31092,10.11.10.103:31093"
	var client = kafka.NewClient(brokers)
	fmt.Println(client.Topics())
	fmt.Println(client.Partitions())
	client.CreateTopic("RezaTopic", 1, 1)

	var producer = client.NewProducer("RezaTopic")

	var c = 0
	for {
		var message = struct {
			Name    string
			Counter int
		}{
			Name:    "Greg!",
			Counter: c,
		}
		c++

		producer.Write(kafka.Message{
			Value: []byte(text.ToJSON(message)),
		})
		time.Sleep(1 * time.Second)
	}
}
