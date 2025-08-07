package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go_code_reviewer/pkg/log"
)

const defaultTimeoutMeiliSeconds = 2000

type kafkaConsumer struct {
	conf         ConsumerConfig
	client       *kafka.Consumer
	messagesChan chan *kafka.Message
}

func NewConsumer(conf ConsumerConfig) (Consumer, error) {
	client, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  conf.Brokers,
		"group.id":           conf.GroupID,
		"auto.offset.reset":  conf.AutoOffset,
		"enable.auto.commit": false,
	})
	if err != nil {
		log.GetLogger().WithError(err).Fatal("failed to connect to kafka")
	}

	if err := client.SubscribeTopics(conf.Topics, nil); err != nil {
		log.GetLogger().WithError(err).Fatal("failed to subscribe to topics")
	}

	return &kafkaConsumer{
		conf:         conf,
		client:       client,
		messagesChan: make(chan *kafka.Message),
	}, nil
}

func (c *kafkaConsumer) Start() error {
	for {
		ev := c.client.Poll(defaultTimeoutMeiliSeconds)
		switch e := ev.(type) {
		case *kafka.Message:
			if e.TopicPartition.Error != nil {
				return e.TopicPartition.Error
			}
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
