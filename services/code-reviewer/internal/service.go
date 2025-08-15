package internal

import (
	"context"
	"errors"
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromaembedding "github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
	"github.com/google/go-github/v58/github"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	langchainopenai "github.com/tmc/langchaingo/llms/openai"
	"go_code_reviewer/pkg/kafka"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/code-reviewer/internal/assistant"
	"go_code_reviewer/services/code-reviewer/internal/config"
	"go_code_reviewer/services/code-reviewer/internal/embedder"
	eventprocessor "go_code_reviewer/services/code-reviewer/internal/event-processor"
	"go_code_reviewer/services/code-reviewer/internal/metrics"
	"go_code_reviewer/services/code-reviewer/internal/parser"
	"go_code_reviewer/services/code-reviewer/internal/repositories"
	"go_code_reviewer/services/code-reviewer/internal/vsc"
	"golang.org/x/oauth2"
)

type Service struct {
	embeddingClient embedder.EmbeddingClient
	llm             llms.Model
	chromaClient    chroma.Client
	vscClient       vsc.VersionControlSystem
	kafkaConsumer   kafka.Consumer
}

func (s *Service) Start() {
	logger := log.GetLogger()
	serviceConfig, err := config.LoadConfig("./config.yaml")
	if err != nil {
		logger.WithError(err).Fatal("failed to load config.yaml")
	}

	logger.WithFields(logrus.Fields{
		"llm_model":       serviceConfig.LLM.Model,
		"llm_temperature": serviceConfig.LLM.Temperature,
		"llm_max_tokens":  serviceConfig.LLM.MaxTokens,
		"llm_base_url":    serviceConfig.LLM.APIBaseURL,
		"embedding_model": serviceConfig.Embedding.Model,
	}).Info("Config loaded for code reviewer")

	if err := s.ConnectToServices(serviceConfig); err != nil {
		logger.WithError(err).Fatal("failed to connect to services")
	}

	openaiEmbeddingFunc, err := chromaembedding.NewOpenAIEmbeddingFunction(
		serviceConfig.LLM.OpenApiKey,
		chromaembedding.WithBaseURL(serviceConfig.Embedding.APIBaseURL),
		chromaembedding.WithModel(chromaembedding.EmbeddingModel(serviceConfig.Embedding.Model)),
	)
	if err != nil {
		logger.WithError(err).Fatal("failed to create openai embedding function")
	}

	embeddingsRepo := repositories.NewEmbeddingRepository(s.chromaClient, openaiEmbeddingFunc, serviceConfig.ChromaDB.CollectionName)
	projectParser := parser.NewProjectParser(map[string]*parser.CodeParser{
		".py": parser.NewCodeParser(parser.LanguagePython),
		".go": parser.NewCodeParser(parser.LanguageGo),
	})

	projectEmbedder := embedder.NewProjectEmbedder(s.embeddingClient, embeddingsRepo, serviceConfig.Embedding.Model)
	codeAssistant := assistant.NewAssistant(serviceConfig, embeddingsRepo, s.llm, s.embeddingClient)
	eventProcessor := eventprocessor.NewModule(projectParser, projectEmbedder, codeAssistant, s.vscClient, s.kafkaConsumer, serviceConfig.WorkerCount)

	eventProcessor.Start()
	err = s.kafkaConsumer.Start()
	if err != nil {
		logger.WithError(err).Fatal("failed to start kafka consumer")
	}
}

func (s *Service) Close() {
	s.chromaClient.Close()
	s.kafkaConsumer.Close()
}

func (s *Service) ConnectToServices(serviceConfig *config.Config) error {
	// connect to embedding client
	embeddingClientConfig := openai.DefaultConfig(serviceConfig.LLM.OpenApiKey)
	if serviceConfig.Embedding.APIBaseURL == "" {
		return errors.New("empty embedding api base url")
	}
	embeddingClientConfig.BaseURL = serviceConfig.Embedding.APIBaseURL
	s.embeddingClient = embedder.NewOpenAiEmbeddingClient(openai.NewClientWithConfig(embeddingClientConfig))

	// connect to llm client
	llm, err := langchainopenai.New(langchainopenai.WithBaseURL(serviceConfig.LLM.APIBaseURL), langchainopenai.WithModel(serviceConfig.LLM.Model), langchainopenai.WithToken(serviceConfig.LLM.OpenApiKey))
	if err != nil {
		return err
	}
	s.llm = llm

	// connect to chroma db client
	chromaClient, err := chroma.NewHTTPClient(chroma.WithBaseURL(serviceConfig.ChromaDB.Address))
	if err != nil {
		return err
	}
	s.chromaClient = chromaClient

	// connect to github
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: serviceConfig.Github.AccessToken})
	tc := oauth2.NewClient(context.Background(), ts)
	s.vscClient = vsc.NewGithub(github.NewClient(tc))

	// connect to prometheus
	metrics.Init(serviceConfig.Prometheus.Address)

	// connect to kafka
	s.kafkaConsumer, err = kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:    serviceConfig.Kafka.Brokers,
		GroupID:    serviceConfig.Kafka.GroupID,
		Topics:     []string{serviceConfig.Kafka.Topics},
		AutoOffset: serviceConfig.Kafka.AutoOffset,
	}, kafka.WithMetricsHandler(metrics.Get().ObserveKafkaPublish))
	if err != nil {
		return err
	}

	return nil
}
