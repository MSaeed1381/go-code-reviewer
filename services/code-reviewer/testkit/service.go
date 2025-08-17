package testkit

import (
	"go.uber.org/mock/gomock"
	kafkamocks "go_code_reviewer/pkg/kafka/mocks"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/api-gateway/pkg/models"
	"go_code_reviewer/services/code-reviewer/internal/assistant"
	"go_code_reviewer/services/code-reviewer/internal/config"
	"go_code_reviewer/services/code-reviewer/internal/embedder"
	embeddermock "go_code_reviewer/services/code-reviewer/internal/embedder/mocks"
	eventprocessor "go_code_reviewer/services/code-reviewer/internal/event-processor"
	"go_code_reviewer/services/code-reviewer/internal/mocks"
	"go_code_reviewer/services/code-reviewer/internal/parser"
	"go_code_reviewer/services/code-reviewer/internal/repositories"
	repositoriesmock "go_code_reviewer/services/code-reviewer/internal/repositories/mocks"
	vscmock "go_code_reviewer/services/code-reviewer/internal/vsc/mocks"
	"testing"
)

type Service struct {
	EmbeddingClient  *embeddermock.MockEmbeddingClient
	LLM              *mocks.MockModel
	ChromaClient     *mocks.MockClient
	VSCClient        *vscmock.MockVersionControlSystem
	KafkaConsumer    *kafkamocks.MockConsumer
	EmbeddingRepo    *repositoriesmock.MockEmbeddingsRepository
	ChromaCollection *mocks.MockCollection
}

func NewService(t *testing.T) *Service {
	controller := gomock.NewController(t)
	return &Service{
		EmbeddingClient:  embeddermock.NewMockEmbeddingClient(controller),
		LLM:              mocks.NewMockModel(controller),
		ChromaClient:     mocks.NewMockClient(controller),
		VSCClient:        vscmock.NewMockVersionControlSystem(controller),
		KafkaConsumer:    kafkamocks.NewMockConsumer(controller),
		EmbeddingRepo:    repositoriesmock.NewMockEmbeddingsRepository(controller),
		ChromaCollection: mocks.NewMockCollection(controller),
	}
}

func (s *Service) Start() {
	logger := log.GetLogger()
	serviceConfig, err := config.LoadConfig("./config.yaml")
	if err != nil {
		logger.WithError(err).Fatal("failed to load config.yaml")
	}

	embeddingsRepo := repositories.NewEmbeddingRepository(s.ChromaClient, nil, serviceConfig.ChromaDB.CollectionName)
	projectParser := parser.NewProjectParser(map[string]*parser.CodeParser{
		".py": parser.NewCodeParser(parser.LanguagePython),
		".go": parser.NewCodeParser(parser.LanguageGo),
	})

	projectEmbedder := embedder.NewProjectEmbedder(s.EmbeddingClient, embeddingsRepo, serviceConfig.Embedding.Model)
	codeAssistant := assistant.NewAssistant(serviceConfig, embeddingsRepo, s.LLM, s.EmbeddingClient)
	eventProcessor := eventprocessor.NewModule(projectParser, projectEmbedder, codeAssistant, s.VSCClient, s.KafkaConsumer, serviceConfig.WorkerCount)

	eventProcessor.Start()
	err = s.KafkaConsumer.Start()
	if err != nil {
		logger.WithError(err).Fatal("failed to start kafka consumer")
	}
}

func GenerateRandomPullRequestEvent() *models.PullRequestEvent {
	return &models.PullRequestEvent{
		Owner:    "MSaeed1381",
		Repo:     "message-broker",
		Number:   51,
		CloneURL: "https://github.com/MSaeed1381/message-broker.git",
		Branch:   "MSaeed1381-patch-52",
		Title:    "Update main.go",
		Author:   "MSaeed1381",
		DiffURL:  "https://github.com/MSaeed1381/message-broker/pull/51.diff",
	}
}
