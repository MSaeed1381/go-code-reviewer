package internal

import (
	"context"
	"errors"
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	openai_embedder "github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"go_code_reviewer/api"
	"go_code_reviewer/internal/assistant"
	"go_code_reviewer/internal/config"
	"go_code_reviewer/internal/embedder"
	"go_code_reviewer/internal/modules"
	"go_code_reviewer/internal/parser"
	"go_code_reviewer/pkg/log"
	"net/http"
)

type ServiceInterface interface {
	Start()
	Close()
}

type Service struct {
	httpServer *http.Server
}

func (service *Service) Start() {
	serviceConfig, err := config.LoadConfig("./config.yaml")
	if err != nil {
		logrus.WithError(err).Error("failed to load config.yaml")
	}

	log.Init(log.Config{
		Level:     logrus.InfoLevel,
		Env:       serviceConfig.Env,
		LogToFile: true,
		FilePath:  "./service.log",
	})
	logger := log.GetLogger()

	logger.WithFields(logrus.Fields{
		"llm_model":       serviceConfig.LLM.Model,
		"llm_temperature": serviceConfig.LLM.Temperature,
		"llm_max_tokens":  serviceConfig.LLM.MaxTokens,
		"llm_base_url":    serviceConfig.LLM.APIBaseURL,
		"embedding_model": serviceConfig.Embedding.Model,
	}).Info("Config loaded for code reviewer service")

	clientConfig := openai.DefaultConfig(serviceConfig.LLM.OpenApiKey)
	clientConfig.BaseURL = serviceConfig.Embedding.APIBaseURL
	embeddingClient := openai.NewClientWithConfig(clientConfig)
	projectEmbedder, err := embedder.NewProjectEmbedder(serviceConfig, embeddingClient)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create project embedder")
		return
	}

	llmClientConfig := openai.DefaultConfig(serviceConfig.LLM.OpenApiKey)
	if serviceConfig.LLM.APIBaseURL != "" {
		llmClientConfig.BaseURL = serviceConfig.LLM.APIBaseURL
	}
	llmClient := openai.NewClientWithConfig(llmClientConfig)

	chromaClient, err := chroma.NewHTTPClient(chroma.WithBaseURL(serviceConfig.ChromaDB.Address))
	if err != nil {
		logger.Fatalf("Failed to create chroma client for assistant: %v", err)
	}

	openaiEmbeddingFunc, err := openai_embedder.NewOpenAIEmbeddingFunction(
		serviceConfig.LLM.OpenApiKey,
		openai_embedder.WithBaseURL("https://api.metisai.ir/openai/v1"),
		openai_embedder.WithModel("text-embedding-3-small"),
	)
	if err != nil {
		logger.Fatalf("Failed to create embedding function: %v", err)
	}

	chromaCollection, err := chromaClient.GetCollection(context.Background(), "coderag", chroma.WithEmbeddingFunctionGet(openaiEmbeddingFunc))
	if err != nil {
		logger.Fatalf("Failed to get collection for assistant: %v", err)
	}

	codeAssistant := assistant.NewAssistant(serviceConfig, chromaCollection, llmClient)
	projectParser := parser.NewProjectParser()

	module := modules.NewModule(projectParser, projectEmbedder, codeAssistant)
	handler := api.NewHandler(module)
	r := handler.RegisterRoutes()

	service.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := service.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Fatalf("failed to start")
	}
}

func (service *Service) Close() {
	service.httpServer.Shutdown(context.Background())
}
