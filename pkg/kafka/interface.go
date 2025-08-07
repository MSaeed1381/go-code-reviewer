package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer interface {
	Start() error
	Close() error
	CommitMessage(msg *kafka.Message) error
	Channel() chan *kafka.Message
}

type Producer interface {
	Send(topic string, value []byte) error
	Close()
}

type ConsumerHandler interface {
	Process(msg *kafka.Message) error
}
