package internal

import (
	"context"
	"errors"
	"github.com/google/go-github/v58/github"
	"go_code_reviewer/pkg/kafka"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/api-gateway/api"
	"go_code_reviewer/services/api-gateway/internal/config"
	eventprocessor "go_code_reviewer/services/api-gateway/internal/event-processor"
	"net/http"
	"time"
)

type Service struct {
	httpServer    *http.Server
	githubClient  *github.Client
	kafkaProducer kafka.Producer
}

func (s *Service) Start() {
	logger := log.GetLogger()
	serviceConfig, err := config.LoadConfig("./config.yaml")
	if err != nil {
		logger.WithError(err).Fatal("failed to load config.yaml")
	}

	err = s.connectToServices(serviceConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to services")
	}

	eventProcessorModule := eventprocessor.New(s.kafkaProducer, serviceConfig.Kafka.Topic)
	handler := api.NewHandler(serviceConfig, eventProcessorModule)
	r := handler.RegisterRoutes()
	s.httpServer = &http.Server{
		Addr:    serviceConfig.HttpServer.Address,
		Handler: r,
	}

	logger.Info("server running on " + serviceConfig.HttpServer.Address)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Fatal("failed to start http server")
	}
}

func (s *Service) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	s.httpServer.Shutdown(ctx)
	s.kafkaProducer.Close()
}

func (s *Service) connectToServices(serviceConfig *config.Config) error {
	// connect to kafka
	kafkaProducer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: serviceConfig.Kafka.Brokers,
	})
	if err != nil {
		return err
	}
	s.kafkaProducer = kafkaProducer
	return nil
}
