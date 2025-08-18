package event_sender

import (
	"context"
	"encoding/json"
	"go_code_reviewer/pkg/kafka"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/pkg/retry"
	"go_code_reviewer/services/api-gateway/pkg/models"
	"time"
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

func (m *Module) ProcessEvent(ctx context.Context, event *models.PullRequestEvent) error {
	logger := log.GetLogger()

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.WithError(err).Error("failed to marshal event")
		return err
	}

	retrier := retry.New[bool](retry.Options{
		MaxRetries: 3,
		Strategy:   retry.ExponentialBackoff(time.Second),
	})
	_, err = retrier.Do(ctx, func() (bool, error) {
		err = m.producer.Send(m.eventTopic, eventBytes)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		logger.WithError(err).Error("failed to send event to kafka")
		return err
	}

	return nil
}
