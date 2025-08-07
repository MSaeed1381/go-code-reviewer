package kafka

type ConsumerConfig struct {
	Brokers    string
	GroupID    string
	Topics     []string
	AutoOffset string
}

type ProducerConfig struct {
	Brokers string
}
