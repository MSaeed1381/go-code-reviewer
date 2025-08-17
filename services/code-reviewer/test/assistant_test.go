package test

import (
	"context"
	"errors"
	"github.com/bmizerany/assert"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/mock/gomock"
	"go_code_reviewer/services/code-reviewer/internal/assistant"
	"go_code_reviewer/services/code-reviewer/internal/config"
	"go_code_reviewer/services/code-reviewer/internal/embedder"
	mockembedder "go_code_reviewer/services/code-reviewer/internal/embedder/mocks"
	"go_code_reviewer/services/code-reviewer/internal/mocks"
	"go_code_reviewer/services/code-reviewer/internal/models"
	mockrepositories "go_code_reviewer/services/code-reviewer/internal/repositories/mocks"
	"testing"
)

func TestAssistant_PerformTask_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepositories.NewMockEmbeddingsRepository(ctrl)
	mockEmbeddingClient := mockembedder.NewMockEmbeddingClient(ctrl)
	mockLLM := mocks.NewMockModel(ctrl)

	ctx := context.Background()
	queryText := "is this function correct?"
	projectID := "proj-1"
	cfg := &config.Config{
		Tasks: config.TasksSection{
			CodeReview: config.TaskConfig{
				Prompts: config.PromptSection{ZeroShot: "Review this: {{.text}} with context: {{.context}}"},
			},
		},
		LLM: config.LLMSection{MaxTokens: 100},
	}

	assistantModule := assistant.NewAssistant(cfg, mockRepo, mockLLM, mockEmbeddingClient)
	mockEmbeddingClient.EXPECT().
		CreateEmbeddings(ctx, string(openai.SmallEmbedding3), []string{queryText}).
		Return([]embedder.Embedding{{Embedding: []float32{0.1, 0.2}}}, nil)

	mockRepo.EXPECT().
		GetNearestRecord(ctx, []float32{0.1, 0.2}, 5, projectID).
		Return([]*models.Snippet{{Content: "func main() {}", Filename: "main.go"}}, nil)

	mockLLM.EXPECT().
		GenerateContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&llms.ContentResponse{
			Choices: []*llms.ContentChoice{{Content: "The code looks good!"}}}, nil).Times(1)

	response, err := assistantModule.PerformTask(ctx, assistant.TaskCodeReview, queryText, projectID)

	require.NoError(t, err)
	assert.Equal(t, "The code looks good!", response)
}

func TestAssistant_PerformTask_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepositories.NewMockEmbeddingsRepository(ctrl)
	mockEmbeddingClient := mockembedder.NewMockEmbeddingClient(ctrl)
	mockLLM := mocks.NewMockModel(ctrl)
	cfg := &config.Config{}

	assistantModule := assistant.NewAssistant(cfg, mockRepo, mockLLM, mockEmbeddingClient)
	ctx := context.Background()
	queryText := "test query"
	projectID := "proj-1"
	expectedErr := errors.New("database failed")

	mockEmbeddingClient.EXPECT().CreateEmbeddings(gomock.Any(), gomock.Any(), gomock.Any()).Return([]embedder.Embedding{{}}, nil)
	mockRepo.EXPECT().GetNearestRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	_, err := assistantModule.PerformTask(ctx, assistant.TaskCodeReview, queryText, projectID)
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
