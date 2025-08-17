package test

import (
	"context"
	"errors"
	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go_code_reviewer/services/code-reviewer/internal/embedder"
	mockembedder "go_code_reviewer/services/code-reviewer/internal/embedder/mocks"
	"go_code_reviewer/services/code-reviewer/internal/models"
	mockrepositories "go_code_reviewer/services/code-reviewer/internal/repositories/mocks"
	"testing"
)

func TestProjectEmbedder_EmbedProject_Success(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockEmbeddingClient := mockembedder.NewMockEmbeddingClient(controller)
	mockEmbeddingsRepo := mockrepositories.NewMockEmbeddingsRepository(controller)

	projectEmbedder := embedder.NewProjectEmbedder(mockEmbeddingClient, mockEmbeddingsRepo, "text-embedding-ada-002")

	ctx := context.Background()
	projectID := "project-123"
	snippets := []*models.Snippet{
		{ID: "snippet-1", Content: "func main() { fmt.Println(sum(1, 3)) }"},
		{ID: "snippet-2", Content: "func sum(a, b int) int { return a + b })"},
	}
	texts := []string{"func main() { fmt.Println(sum(1, 3)) }", "func sum(a, b int) int { return a + b })"}

	expectedEmbeddings := []embedder.Embedding{
		{Embedding: []float32{0.1, 0.2, 0.3}},
		{Embedding: []float32{0.4, 0.5, 0.6}},
	}

	mockEmbeddingClient.EXPECT().
		CreateEmbeddings(ctx, "text-embedding-ada-002", texts).
		Return(expectedEmbeddings, nil).
		Times(1)

	expectedSnippetsToSave := []*models.Snippet{
		{ID: "snippet-1", Content: "func main() { fmt.Println(sum(1, 3)) }", Embedding: []float32{0.1, 0.2, 0.3}},
		{ID: "snippet-2", Content: "func sum(a, b int) int { return a + b })", Embedding: []float32{0.4, 0.5, 0.6}},
	}
	mockEmbeddingsRepo.EXPECT().
		Add(ctx, expectedSnippetsToSave, projectID).
		Return(nil).
		Times(1)

	err := projectEmbedder.EmbedProject(ctx, projectID, snippets)
	require.NoError(t, err)
	require.Equal(t, expectedEmbeddings[0].Embedding, snippets[0].Embedding)
	require.Equal(t, expectedEmbeddings[1].Embedding, snippets[1].Embedding)
}

func TestProjectEmbedder_EmbedProject_EmbeddingClientError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockEmbeddingClient := mockembedder.NewMockEmbeddingClient(controller)
	mockEmbeddingsRepo := mockrepositories.NewMockEmbeddingsRepository(controller)

	projectEmbedder := embedder.NewProjectEmbedder(mockEmbeddingClient, mockEmbeddingsRepo, "text-embedding-ada-002")

	ctx := context.Background()
	projectID := "project-123"
	snippets := []*models.Snippet{
		{ID: "snippet-1", Content: "func main() {}"},
	}
	texts := []string{"func main() {}"}
	expectedError := errors.New("embedding error")

	mockEmbeddingClient.EXPECT().
		CreateEmbeddings(ctx, "text-embedding-ada-002", texts).
		Return(nil, expectedError).
		Times(1)

	mockEmbeddingsRepo.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	err := projectEmbedder.EmbedProject(ctx, projectID, snippets)
	require.Error(t, err)
	require.Equal(t, expectedError, err)
}

func TestProjectEmbedder_EmbedProject_RepositoryError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockEmbeddingClient := mockembedder.NewMockEmbeddingClient(controller)
	mockEmbeddingsRepo := mockrepositories.NewMockEmbeddingsRepository(controller)

	projectEmbedder := embedder.NewProjectEmbedder(mockEmbeddingClient, mockEmbeddingsRepo, "text-embedding-ada-002")

	ctx := context.Background()
	projectID := "project-123"
	snippets := []*models.Snippet{
		{ID: "snippet-1", Content: "func main() {}"},
	}
	texts := []string{"func main() {}"}
	expectedEmbeddings := []embedder.Embedding{
		{Embedding: []float32{0.1, 0.2, 0.3}},
	}
	expectedError := errors.New("database connection failed")

	mockEmbeddingClient.EXPECT().
		CreateEmbeddings(ctx, "text-embedding-ada-002", texts).
		Return(expectedEmbeddings, nil).
		Times(1)

	expectedSnippetsToSave := []*models.Snippet{
		{ID: "snippet-1", Content: "func main() {}", Embedding: []float32{0.1, 0.2, 0.3}},
	}
	mockEmbeddingsRepo.EXPECT().
		Add(ctx, expectedSnippetsToSave, projectID).
		Return(expectedError).
		Times(1)

	err := projectEmbedder.EmbedProject(ctx, projectID, snippets)
	require.Error(t, err)
	assert.Equal(t, expectedError, err)
}
