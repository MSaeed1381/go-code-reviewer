package internal

import (
	"context"
	"errors"
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	openaiembedder "github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"go_code_reviewer/api"
	"go_code_reviewer/internal/assistant"
	"go_code_reviewer/internal/code_reviewer"
	"go_code_reviewer/internal/config"
	"go_code_reviewer/internal/embedder"
	"go_code_reviewer/internal/parser"
	"go_code_reviewer/pkg/log"
	"net/http"
	"time"
)

type Service struct {
	httpServer       *http.Server
	embeddingClient  *openai.Client
	llmClient        *openai.Client
	chromaCollection chroma.Collection
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
	}).Info("Config loaded for code reviewer s")

	if err := s.ConnectToServices(serviceConfig); err != nil {
		logger.WithError(err).Fatal("failed to connect to services")
	}

	projectParser := parser.NewProjectParser(map[string]*parser.CodeParser{
		".py": parser.NewCodeParser(parser.LanguagePython),
		".go": parser.NewCodeParser(parser.LanguageGo),
	})
	projectEmbedder := embedder.NewProjectEmbedder(s.embeddingClient, s.chromaCollection, openai.EmbeddingModel(serviceConfig.Embedding.Model))
	codeAssistant := assistant.NewAssistant(serviceConfig, s.chromaCollection, s.llmClient)
	codeReviewerModule := code_reviewer.NewModule(projectParser, projectEmbedder, codeAssistant)

	handler := api.NewHandler(codeReviewerModule)
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
	s.chromaCollection.Close()
}

func (s *Service) ConnectToServices(serviceConfig *config.Config) error {
	// connect to embedding client
	embeddingClientConfig := openai.DefaultConfig(serviceConfig.LLM.OpenApiKey)
	if serviceConfig.Embedding.APIBaseURL == "" {
		return errors.New("empty embedding api base url")
	}
	embeddingClientConfig.BaseURL = serviceConfig.Embedding.APIBaseURL
	s.embeddingClient = openai.NewClientWithConfig(embeddingClientConfig)

	// connect to llm client
	llmClientConfig := openai.DefaultConfig(serviceConfig.LLM.OpenApiKey)
	if serviceConfig.LLM.APIBaseURL == "" {
		return errors.New("empty llm api base url")
	}
	llmClientConfig.BaseURL = serviceConfig.LLM.APIBaseURL
	s.llmClient = openai.NewClientWithConfig(llmClientConfig)

	// connect to chroma db client
	chromaClient, err := chroma.NewHTTPClient(chroma.WithBaseURL(serviceConfig.ChromaDB.Address))
	if err != nil {
		return err
	}
	defer chromaClient.Close()

	openaiEmbeddingFunc, err := openaiembedder.NewOpenAIEmbeddingFunction(
		serviceConfig.LLM.OpenApiKey,
		openaiembedder.WithBaseURL(serviceConfig.Embedding.APIBaseURL),
		openaiembedder.WithModel(openaiembedder.EmbeddingModel(serviceConfig.Embedding.Model)),
	)
	if err != nil {
		return err
	}

	// connect to chroma db collection
	chromaCollection, err := chromaClient.GetCollection(context.Background(), serviceConfig.ChromaDB.CollectionName, chroma.WithEmbeddingFunctionGet(openaiEmbeddingFunc))
	if err != nil {
		return err
	}
	s.chromaCollection = chromaCollection

	return nil
}
