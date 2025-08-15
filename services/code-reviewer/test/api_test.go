package test

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/mock/gomock"
	"go_code_reviewer/services/api-gateway/pkg/models"
	"go_code_reviewer/services/code-reviewer/internal/embedder"
	"go_code_reviewer/services/code-reviewer/internal/mocks"
	"go_code_reviewer/services/code-reviewer/testkit"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProcessPullRequestEvent(t *testing.T) {
	service := testkit.NewService(t)
	service.ChromaClient.EXPECT().GetOrCreateCollection(gomock.Any(), "coderag").Return(service.ChromaCollection, nil).Times(1)
	service.KafkaConsumer.EXPECT().Start().Times(1)
	ch := make(chan *kafka.Message, 1)
	service.KafkaConsumer.EXPECT().Channel().Return(ch).AnyTimes()

	llmReview := "reject"
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
	diffContent := `"+fmt.Println("Hello, Word!")\n-fmt.Println("Hello, World!")"`
	diffEmbedding := []float32{2, 4, 2}
	mainEmbedding := []float32{1, 2, 4}

	marshal, err := json.Marshal(prEvent)
	require.NoError(t, err)

	kafkaMessage := &kafka.Message{
		Value: marshal,
	}
	ch <- kafkaMessage

	dirPath, err := os.MkdirTemp("", "gh-pr-*")
	require.NoError(t, err)

	cleanup := func() error {
		return os.RemoveAll(dirPath)
	}

	filePath := filepath.Join(dirPath, "main.go")
	goCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`
	err = os.WriteFile(filePath, []byte(goCode), 0644)
	require.NoError(t, err)
	service.VSCClient.EXPECT().Clone(gomock.Any(), prEvent.CloneURL, prEvent.Branch).Return(dirPath, cleanup, nil).Times(1)
	service.EmbeddingClient.EXPECT().CreateEmbeddings(gomock.Any(), gomock.Any(), []string{`func main() {
	fmt.Println("Hello, World!")
}`}).Return([]embedder.Embedding{{
		Embedding: mainEmbedding,
	}}, nil).Times(1)
	service.EmbeddingClient.EXPECT().CreateEmbeddings(gomock.Any(), gomock.Any(), []string{diffContent}).Return([]embedder.Embedding{{
		Embedding: diffEmbedding,
	}}, nil).Times(1)

	service.ChromaCollection.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	service.VSCClient.EXPECT().DownloadUrl(gomock.Any(), prEvent.DiffURL).Return(diffContent, nil)

	queryResult := mocks.NewMockQueryResult(gomock.NewController(t))
	service.ChromaCollection.EXPECT().Query(gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).Return(queryResult, nil).Times(1)

	queryResult.EXPECT().GetDocumentsGroups().Times(1)
	service.LLM.EXPECT().GenerateContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: llmReview}}}, nil).Times(1)
	service.VSCClient.EXPECT().PostPRComment(gomock.Any(), prEvent.Number, llmReview, prEvent.Owner, prEvent.Repo).Return(nil).Times(1)
	service.KafkaConsumer.EXPECT().CommitMessage(kafkaMessage).Return(nil).Times(1)

	service.Start()
	time.Sleep(1 * time.Second)
}
