package internal

import (
	"encoding/json"
	"fmt"
	"go_code_reviewer/pkg/kafka"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/api-gateway/pkg/models"
	"go_code_reviewer/services/code-reviewer/cmd/load-test/internal/config"
	"time"
)

type LoadTestModule struct {
	config   config.LoadTestConfig
	producer kafka.Producer
}

func NewLoadTestModule(config config.LoadTestConfig, producer kafka.Producer) *LoadTestModule {
	return &LoadTestModule{
		config:   config,
		producer: producer,
	}
}

func (m *LoadTestModule) Start() {
	logEntry := log.GetLogger()
	ticker := time.NewTicker(time.Second / time.Duration(m.config.MessagePerSecond))
	defer ticker.Stop()

	prEvent := &models.PullRequestEvent{
		Owner:    "MSaeed1381",
		Repo:     "message-broker",
		Number:   51,
		CloneURL: "https://github.com/MSaeed1381/message-broker.git",
		Branch:   "MSaeed1381-patch-52",
		Title:    "Update main.go",
		Author:   "MSaeed1381",
		DiffURL:  "https://github.com/MSaeed1381/message-broker/pull/51.diff",
	}
	marshal, err := json.Marshal(prEvent)
	if err != nil {
		logEntry.WithError(err).Fatal("failed to marshal kafka event")
	}

	i := 0
	for {
		select {
		case <-ticker.C:
			go func() {
				m.producer.Send(m.config.KafkaTopic, marshal)
				fmt.Println(i)
				i = i + 1
			}()
		}
	}
}
