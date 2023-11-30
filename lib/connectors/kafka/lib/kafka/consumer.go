package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"strings"
	"time"
)

type Consumer struct {
	Config          *kafka.ReaderConfig
	Reader          *kafka.Reader
	MessageHandlers []ConsumerHandler
	ErrorHandler    func(err error) bool
	Context         context.Context
}

type ConsumerHandler func(message Message)

type ConsumerConfig struct {
	// The list of broker addresses used to connect to the kafka cluster.
	_Brokers *[]string

	// GroupID holds the optional consumer group id.  If GroupID is specified, then
	// Partition should NOT be specified e.g. 0
	_GroupID *string

	// GroupTopics allows specifying multiple topics, but can only be used in
	// combination with GroupID, as it is a consumer-group feature. As such, if
	// GroupID is set, then either Topic or GroupTopics must be defined.
	_GroupTopics *[]string

	// The topic to read messages from.
	_Topic *string

	// Partition to read messages from.  Either Partition or GroupID may
	// be assigned, but not both
	_Partition *int

	// The capacity of the internal message queue, defaults to 100 if none is
	// set.
	_QueueCapacity *int

	// MinBytes indicates to the broker the minimum batch size that the consumer
	// will accept. Setting a high minimum when consuming from a low-volume topic
	// may result in delayed delivery when the broker does not have enough data to
	// satisfy the defined minimum.
	//
	// Default: 1
	_MinBytes *int

	// MaxBytes indicates to the broker the maximum batch size that the consumer
	// will accept. The broker will truncate a message to satisfy this maximum, so
	// choose a value that is high enough for your largest message size.
	//
	// Default: 1MB
	_MaxBytes *int

	// Maximum amount of time to wait for new data to come when fetching batches
	// of messages from kafka.
	//
	// Default: 10s
	_MaxWait *time.Duration

	// ReadLagInterval sets the frequency at which the reader lag is updated.
	// Setting this field to a negative value disables lag reporting.
	_ReadLagInterval *time.Duration

	// HeartbeatInterval sets the optional frequency at which the reader sends the consumer
	// group heartbeat update.
	//
	// Default: 3s
	//
	// Only used when GroupID is set
	_HeartbeatInterval *time.Duration

	// CommitInterval indicates the interval at which offsets are committed to
	// the broker.  If 0, commits will be handled synchronously.
	//
	// Default: 0
	//
	// Only used when GroupID is set
	_CommitInterval *time.Duration

	// PartitionWatchInterval indicates how often a reader checks for partition changes.
	// If a reader sees a partition change (such as a partition add) it will rebalance the group
	// picking up new partitions.
	//
	// Default: 5s
	//
	// Only used when GroupID is set and WatchPartitionChanges is set.
	_PartitionWatchInterval *time.Duration

	// WatchForPartitionChanges is used to inform kafka-go that a consumer group should be
	// polling the brokers and rebalancing if any partition changes happen to the topic.
	_WatchPartitionChanges *bool

	// SessionTimeout optionally sets the length of time that may pass without a heartbeat
	// before the coordinator considers the consumer dead and initiates a rebalance.
	//
	// Default: 30s
	//
	// Only used when GroupID is set
	_SessionTimeout *time.Duration

	// RebalanceTimeout optionally sets the length of time the coordinator will wait
	// for members to join as part of a rebalance.  For kafka servers under higher
	// load, it may be useful to set this value higher.
	//
	// Default: 30s
	//
	// Only used when GroupID is set
	_RebalanceTimeout *time.Duration

	// JoinGroupBackoff optionally sets the length of time to wait between re-joining
	// the consumer group after an error.
	//
	// Default: 5s
	_JoinGroupBackoff *time.Duration

	// RetentionTime optionally sets the length of time the consumer group will be saved
	// by the broker
	//
	// Default: 24h
	//
	// Only used when GroupID is set
	_RetentionTime *time.Duration

	// StartOffset determines from whence the consumer group should begin
	// consuming when it finds a partition without a committed offset.  If
	// non-zero, it must be set to one of FirstOffset or LastOffset.
	//
	// Default: FirstOffset
	//
	// Only used when GroupID is set
	_StartOffset *int64

	// BackoffDelayMin optionally sets the smallest amount of time the reader will wait before
	// polling for new messages
	//
	// Default: 100ms
	_ReadBackoffMin *time.Duration

	// BackoffDelayMax optionally sets the maximum amount of time the reader will wait before
	// polling for new messages
	//
	// Default: 1s
	_ReadBackoffMax *time.Duration

	// Limit of how many attempts will be made before delivering the error.
	//
	// The default is to try 3 times.
	_MaxAttempts *int
}

func NewConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{}
}

func NewConsumer(servers, topic string, configs ...*ConsumerConfig) *Consumer {
	var c = Consumer{}
	c.Config = &kafka.ReaderConfig{
		Brokers: strings.Split(servers, ","),
		Topic:   topic,
	}
	for _, config := range configs {
		applyConfig(config, c.Config)
	}

	c.Reader = kafka.NewReader(*c.Config)
	c.Context = context.Background()
	go func() {
		for {
			select {
			case <-c.Context.Done():
				c.Reader.Close()
			default:
				msg, err := c.Reader.ReadMessage(context.Background())
				if err != nil {
					if c.ErrorHandler != nil && !c.ErrorHandler(err) {
						return
					}
				} else {
					for _, handler := range c.MessageHandlers {
						handler(Message(msg))
					}
				}
			}

		}
	}()

	return &c
}

func (c *Consumer) OnError(fn func(err error) bool) {
	c.ErrorHandler = fn
}

func (c *Consumer) Close() {
	c.Context.Done()
	return
}

func (c *Consumer) OnMessage(fn ConsumerHandler) {
	c.MessageHandlers = append(c.MessageHandlers, fn)
}

func (c *ConsumerConfig) Brokers(input []string) *ConsumerConfig {
	c._Brokers = &input
	return c
}
func (c *ConsumerConfig) GroupID(input string) *ConsumerConfig {
	c._GroupID = &input
	return c
}
func (c *ConsumerConfig) GroupTopics(input []string) *ConsumerConfig {
	c._GroupTopics = &input
	return c
}
func (c *ConsumerConfig) Topic(input string) *ConsumerConfig {
	c._Topic = &input
	return c
}
func (c *ConsumerConfig) Partition(input int) *ConsumerConfig {
	c._Partition = &input
	return c
}
func (c *ConsumerConfig) QueueCapacity(input int) *ConsumerConfig {
	c._QueueCapacity = &input
	return c
}
func (c *ConsumerConfig) MinBytes(input int) *ConsumerConfig {
	c._MinBytes = &input
	return c
}
func (c *ConsumerConfig) MaxBytes(input int) *ConsumerConfig {
	c._MaxBytes = &input
	return c
}
func (c *ConsumerConfig) MaxWait(input time.Duration) *ConsumerConfig {
	c._MaxWait = &input
	return c
}
func (c *ConsumerConfig) ReadLagInterval(input time.Duration) *ConsumerConfig {
	c._ReadLagInterval = &input
	return c
}
func (c *ConsumerConfig) HeartbeatInterval(input time.Duration) *ConsumerConfig {
	c._HeartbeatInterval = &input
	return c
}
func (c *ConsumerConfig) CommitInterval(input time.Duration) *ConsumerConfig {
	c._CommitInterval = &input
	return c
}
func (c *ConsumerConfig) PartitionWatchInterval(input time.Duration) *ConsumerConfig {
	c._PartitionWatchInterval = &input
	return c
}
func (c *ConsumerConfig) WatchPartitionChanges(input bool) *ConsumerConfig {
	c._WatchPartitionChanges = &input
	return c
}
func (c *ConsumerConfig) SessionTimeout(input time.Duration) *ConsumerConfig {
	c._SessionTimeout = &input
	return c
}
func (c *ConsumerConfig) RebalanceTimeout(input time.Duration) *ConsumerConfig {
	c._RebalanceTimeout = &input
	return c
}
func (c *ConsumerConfig) JoinGroupBackoff(input time.Duration) *ConsumerConfig {
	c._JoinGroupBackoff = &input
	return c
}
func (c *ConsumerConfig) RetentionTime(input time.Duration) *ConsumerConfig {
	c._RetentionTime = &input
	return c
}
func (c *ConsumerConfig) StartOffset(input int64) *ConsumerConfig {
	c._StartOffset = &input
	return c
}
func (c *ConsumerConfig) ReadBackoffMin(input time.Duration) *ConsumerConfig {
	c._ReadBackoffMin = &input
	return c
}
func (c *ConsumerConfig) ReadBackoffMax(input time.Duration) *ConsumerConfig {
	c._ReadBackoffMax = &input
	return c
}
func (c *ConsumerConfig) MaxAttempts(input int) *ConsumerConfig {
	c._MaxAttempts = &input
	return c
}
