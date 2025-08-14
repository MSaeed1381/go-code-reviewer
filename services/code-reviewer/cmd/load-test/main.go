package main

import (
	"go_code_reviewer/pkg/kafka"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/code-reviewer/cmd/load-test/internal"
	"go_code_reviewer/services/code-reviewer/cmd/load-test/internal/config"
)

func main() {
	messagesPerSecond := 30
	logEntry := log.GetLogger()
	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: "kafka:9092",
	})
	if err != nil {
		logEntry.WithError(err).Fatal("failed to create producer")
		return
	}
	defer producer.Close()

	module := internal.NewLoadTestModule(config.LoadTestConfig{
		MessagePerSecond: int32(messagesPerSecond),
		KafkaTopic:       "pr-events",
	}, producer)

	module.Start()
}
