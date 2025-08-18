package embedder

import (
	"context"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/pkg/retry"
	"go_code_reviewer/services/code-reviewer/internal/models"
	"go_code_reviewer/services/code-reviewer/internal/repositories"
	"time"
)

type ProjectEmbedder struct {
	embeddingsRepo  repositories.EmbeddingsRepository
	embeddingClient EmbeddingClient
	embeddingModel  string
}

func NewProjectEmbedder(embeddingClient EmbeddingClient, embeddingsRepo repositories.EmbeddingsRepository, embeddingModel string) *ProjectEmbedder {
	return &ProjectEmbedder{
		embeddingsRepo:  embeddingsRepo,
		embeddingClient: embeddingClient,
		embeddingModel:  embeddingModel,
	}
}

func (p *ProjectEmbedder) EmbedProject(ctx context.Context, projectId string, snippets []*models.Snippet) error {
	logger := log.GetLogger()
	var texts []string
	for _, snippet := range snippets {
		texts = append(texts, snippet.Content)
	}

	var retrier = retry.New[[]Embedding](retry.Options{
		MaxRetries: 3,
		Strategy:   retry.ExponentialJitterBackoff(500*time.Millisecond, 10*time.Second),
	})
	embeddings, err := retrier.Do(ctx, func() ([]Embedding, error) {
		return p.embeddingClient.CreateEmbeddings(ctx, p.embeddingModel, texts)
	})
	if err != nil {
		logger.WithError(err).Error("failed to create embeddings")
		return err
	}

	for i := range snippets {
		snippets[i].Embedding = embeddings[i].Embedding
	}

	err = p.embeddingsRepo.Add(ctx, snippets, projectId)
	if err != nil {
		logger.WithError(err).Error("failed to persist embeddings")
		return err
	}

	return nil
}
