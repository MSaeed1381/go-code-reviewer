package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go_code_reviewer/pkg/log"
)

const (
	defaultTimeoutMeiliSeconds = 1000
	defaultChannelSize         = 2000
)

type kafkaConsumer struct {
	conf                  ConsumerConfig
	client                *kafka.Consumer
	messagesChan          chan *kafka.Message
	observeConsumeCounter func(status string)
}

type Option func(consumer *kafkaConsumer)

func WithMetricsHandler(handler func(status string)) Option {
	return func(consumer *kafkaConsumer) {
		consumer.observeConsumeCounter = handler
	}
}

func NewConsumer(conf ConsumerConfig, opts ...Option) (Consumer, error) {
	consumer := kafkaConsumer{}
	for _, opt := range opts {
		opt(&consumer)
	}
	client, err := kafka.NewConsumer(&kafka.ConfigMap{
		bootstrapServersKey: conf.Brokers,
		groupIdKey:          conf.GroupID,
		autoOffsetResetKey:  conf.AutoOffset,
		enableAutoCommitKey: false,
	})
	if err != nil {
		log.GetLogger().WithError(err).Fatal("failed to connect to kafka")
	}

	if err := client.SubscribeTopics(conf.Topics, nil); err != nil {
		log.GetLogger().WithError(err).Fatal("failed to subscribe to topics")
	}

	consumer.conf = conf
	consumer.client = client
	consumer.messagesChan = make(chan *kafka.Message, defaultChannelSize)
	return &consumer, nil
}

func (c *kafkaConsumer) Start() error {
	for {
		ev := c.client.Poll(defaultTimeoutMeiliSeconds)
		switch e := ev.(type) {
		case *kafka.Message:
			if e.TopicPartition.Error != nil {
				return e.TopicPartition.Error
			}
			go c.observeConsumeCounter("success")
			c.messagesChan <- e
		case kafka.Error:
			return e
		case nil:
		}
	}
}

func (c *kafkaConsumer) Close() error {
	close(c.messagesChan)
	return c.client.Close()
}

func (c *kafkaConsumer) CommitMessage(msg *kafka.Message) error {
	_, err := c.client.CommitMessage(msg)
	return err
}

func (c *kafkaConsumer) Channel() chan *kafka.Message {
	return c.messagesChan
}
