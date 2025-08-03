package internal

import (
	"context"
	"errors"
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromaembedding "github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
	"github.com/google/go-github/v58/github"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	langchainopenai "github.com/tmc/langchaingo/llms/openai"
	"go_code_reviewer/api"
	"go_code_reviewer/internal/assistant"
	"go_code_reviewer/internal/code_reviewer"
	"go_code_reviewer/internal/config"
	"go_code_reviewer/internal/embedder"
	"go_code_reviewer/internal/parser"
	"go_code_reviewer/internal/repositories"
	"go_code_reviewer/internal/vsc"
	"go_code_reviewer/pkg/log"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

type Service struct {
	httpServer      *http.Server
	embeddingClient *openai.Client
	llm             *langchainopenai.LLM
	chromaClient    chroma.Client
	githubClient    *github.Client
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

	pullRequestEventChannel := make(chan *vsc.PullRequestEvent)
	projectEmbedder := embedder.NewProjectEmbedder(embedder.NewOpenAiEmbeddingClient(s.embeddingClient), embeddingsRepo, serviceConfig.Embedding.Model)
	codeAssistant := assistant.NewAssistant(serviceConfig, embeddingsRepo, s.llm, embedder.NewOpenAiEmbeddingClient(s.embeddingClient))
	codeReviewerModule := code_reviewer.NewModule(projectParser, projectEmbedder, codeAssistant, pullRequestEventChannel, vsc.NewGithub(s.githubClient))
	go codeReviewerModule.Start()

	handler := api.NewHandler(serviceConfig, codeReviewerModule, pullRequestEventChannel)
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
	s.chromaClient.Close()
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
	s.githubClient = github.NewClient(tc)

	return nil
}
