package kafka

const (
	bootstrapServersKey = "bootstrap.servers"
	groupIdKey          = "group.id"
	autoOffsetResetKey  = "auto.offset.reset"
	enableAutoCommitKey = "enable.auto.commit"
)

type ConsumerConfig struct {
	Brokers    string
	GroupID    string
	Topics     []string
	AutoOffset string
}

type ProducerConfig struct {
	Brokers        string
	metricsHandler func(status string)
}
