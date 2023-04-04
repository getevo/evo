package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"strings"
	"time"
)

type Producer struct {
	Writer  *kafka.Writer
	Context context.Context
}
type RequiredAcks kafka.RequiredAcks

const (
	RequireNone RequiredAcks = 0
	RequireOne  RequiredAcks = 1
	RequireAll  RequiredAcks = -1
)

type BalancerStrategy int

const (
	MurMur2    BalancerStrategy = 0
	CRC32      BalancerStrategy = 1
	Hash       BalancerStrategy = 2
	LeastBytes BalancerStrategy = 3
	RoundRobin BalancerStrategy = 4
)

type Compression = kafka.Compression

const (
	Gzip   Compression = compress.Gzip
	Snappy Compression = compress.Snappy
	Lz4    Compression = compress.Lz4
	Zstd   Compression = compress.Zstd
)

type ProducerConfig struct {

	// Compression set the compression codec to be used to compress messages.
	_Compression *Compression

	_Balancer *kafka.Balancer

	_RequiredAcks *kafka.RequiredAcks

	// The balancer used to distribute messages across partitions.
	//
	// The default is to use a round-robin distribution.
	/*	Balancer Balancer*/

	// Limit on how many attempts will be made to deliver a message.
	//
	// The default is to try at most 10 times.
	_MaxAttempts *int

	// Limit on how many messages will be buffered before being sent to a
	// partition.
	//
	// The default is to use a target batch size of 100 messages.
	_BatchSize *int

	// Limit the maximum size of a request in bytes before being sent to
	// a partition.
	//
	// The default is to use a kafka default value of 1048576.
	_BatchBytes *int64

	// Time limit on how often incomplete message batches will be flushed to
	// kafka.
	//
	// The default is to flush at least every second.
	_BatchTimeout *time.Duration

	// Timeout for read operations performed by the Writer.
	//
	// Defaults to 10 seconds.
	_ReadTimeout *time.Duration

	// Timeout for write operation performed by the Writer.
	//
	// Defaults to 10 seconds.
	_WriteTimeout *time.Duration

	/*	// Number of acknowledges from partition replicas required before receiving
		// a response to a produce request, the following values are supported:
		//
		//  RequireNone (0)  fire-and-forget, do not wait for acknowledgements from the
		//  RequireOne  (1)  wait for the leader to acknowledge the writes
		//  RequireAll  (-1) wait for the full ISR to acknowledge the writes
		//
		// Defaults to RequireNone.
		RequiredAcks RequiredAcks*/

	// Setting this flag to true causes the WriteMessages method to never block.
	// It also means that errors are ignored since the caller will not receive
	// the returned value. Use this only if you don't care about guarantees of
	// whether the messages were written to kafka.
	//
	// Defaults to false.
	_Async *bool
}

func NewProducer(servers, topic string, configs ...*ProducerConfig) *Producer {
	var producer = Producer{
		Writer: &kafka.Writer{
			Addr:  kafka.TCP(strings.Split(servers, ",")...),
			Topic: topic,
		},
	}
	for _, config := range configs {
		applyConfig(config, producer.Writer)
	}
	producer.Context = context.Background()
	return &producer
}

/*func (p *Producer)Topics()  {
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()


	partitions, err := conn.ReadPartitions()
	if err != nil {
		panic(err.Error())
	}

	m := map[string]struct{}{}

	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	for k := range m {
		fmt.Println(k)
	}
}*/

func NewProducerConfig() *ProducerConfig {
	return &ProducerConfig{}
}

func (p *Producer) Write(messages ...Message) error {
	var msgs = make([]kafka.Message, len(messages), len(messages))
	for i, msg := range messages {
		msgs[i] = kafka.Message(msg)
	}

	return p.Writer.WriteMessages(p.Context, msgs...)
}

func (p *Producer) WriteWithContext(ctx context.Context, messages ...Message) error {
	var msgs = make([]kafka.Message, len(messages), len(messages))
	for i, msg := range messages {
		msgs[i] = kafka.Message(msg)
	}
	return p.Writer.WriteMessages(p.Context, msgs...)
}

func (c *ProducerConfig) MaxAttempts(input int) *ProducerConfig {
	c._MaxAttempts = &input
	return c
}
func (c *ProducerConfig) BatchSize(input int) *ProducerConfig {
	c._BatchSize = &input
	return c
}
func (c *ProducerConfig) BatchBytes(input int64) *ProducerConfig {
	c._BatchBytes = &input
	return c
}
func (c *ProducerConfig) BatchTimeout(input time.Duration) *ProducerConfig {
	c._BatchTimeout = &input
	return c
}
func (c *ProducerConfig) ReadTimeout(input time.Duration) *ProducerConfig {
	c._ReadTimeout = &input
	return c
}
func (c *ProducerConfig) WriteTimeout(input time.Duration) *ProducerConfig {
	c._WriteTimeout = &input
	return c
}
func (c *ProducerConfig) Async(input bool) *ProducerConfig {
	c._Async = &input
	return c
}

func (c *ProducerConfig) Compression(input Compression) *ProducerConfig {
	c._Compression = &input
	return c
}

func (c *ProducerConfig) Balancer(input BalancerStrategy) *ProducerConfig {
	var balancer kafka.Balancer
	switch input {
	case MurMur2:
		balancer = kafka.Murmur2Balancer{}
	case CRC32:
		balancer = kafka.CRC32Balancer{}
	case Hash:
		balancer = &kafka.Hash{}
	case RoundRobin:
		balancer = &kafka.RoundRobin{}
	case LeastBytes:
		balancer = &kafka.LeastBytes{}
	default:
		panic("invalid balancer")
	}
	c._Balancer = &balancer
	return c
}

func (c *ProducerConfig) Ack(input RequiredAcks) *ProducerConfig {
	v := kafka.RequiredAcks(input)
	c._RequiredAcks = &v
	return c
}
