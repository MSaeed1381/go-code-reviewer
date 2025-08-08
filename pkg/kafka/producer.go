package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type kafkaProducer struct {
	conf   ProducerConfig
	client *kafka.Producer
}

func NewProducer(conf ProducerConfig) (Producer, error) {
	client, err := kafka.NewProducer(&kafka.ConfigMap{
		bootstrapServersKey: conf.Brokers,
	})
	if err != nil {
		return nil, err
	}

	return &kafkaProducer{
		conf:   conf,
		client: client,
	}, nil
}

func (p *kafkaProducer) Send(topic string, value []byte) error {
	return p.client.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)
}

func (p *kafkaProducer) Close() {
	p.client.Flush(5000)
	p.client.Close()
}
