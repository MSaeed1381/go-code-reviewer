package event_processor

import (
	"encoding/json"
	"go_code_reviewer/pkg/kafka"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/api-gateway/pkg/models"
)

type Module struct {
	eventTopic string
	producer   kafka.Producer
}

func New(producer kafka.Producer, eventTopic string) *Module {
	return &Module{
		producer:   producer,
		eventTopic: eventTopic,
	}
}

func (m *Module) ProcessEvent(event *models.PullRequestEvent) error {
	logger := log.GetLogger()
	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.WithError(err).Error("failed to marshal event")
		return err
	}
	err = m.producer.Send(m.eventTopic, eventBytes)
	if err != nil {
		logger.WithError(err).Error("failed to send event to kafka")
		return err
	}
	return nil
}
