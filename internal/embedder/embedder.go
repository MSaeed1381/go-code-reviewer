package embedder

import (
	"context"
	"go_code_reviewer/internal/models"
	"go_code_reviewer/internal/repositories"
	"go_code_reviewer/pkg/log"
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

func (p *ProjectEmbedder) EmbedProject(ctx context.Context, snippets []*models.Snippet) error {
	logger := log.GetLogger()
	var texts []string
	for _, snippet := range snippets {
		texts = append(texts, snippet.Content)
	}

	embeddings, err := p.embeddingClient.CreateEmbeddings(ctx, p.embeddingModel, texts)
	if err != nil {
		logger.WithError(err).Error("failed to create embeddings")
		return err
	}

	for i := range snippets {
		snippets[i].Embedding = embeddings[i].Embedding
	}

	err = p.embeddingsRepo.Add(ctx, snippets)
	if err != nil {
		logger.WithError(err).Error("failed to persist embeddings")
		return err
	}

	return nil
}
