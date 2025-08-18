package embedder

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/pkg/retry"
)

type Embedding struct {
	Embedding []float32
}

type EmbeddingClient interface {
	CreateEmbeddings(ctx context.Context, embeddingModel string, texts []string) ([]Embedding, error)
}

type OpenAiEmbeddingClient struct {
	openaiClient *openai.Client
	retrier      retry.Retrier[openai.EmbeddingResponse]
}

type OpenAiEmbeddingOption func(client *OpenAiEmbeddingClient)

func WithRetrier(retrier retry.Retrier[openai.EmbeddingResponse]) OpenAiEmbeddingOption {
	return func(client *OpenAiEmbeddingClient) {
		client.retrier = retrier
	}
}

func NewOpenAiEmbeddingClient(openaiClient *openai.Client, opts ...OpenAiEmbeddingOption) EmbeddingClient {
	client := &OpenAiEmbeddingClient{
		openaiClient: openaiClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (e *OpenAiEmbeddingClient) CreateEmbeddings(ctx context.Context, embeddingModel string, texts []string) ([]Embedding, error) {
	logger := log.GetLogger()
	req := openai.EmbeddingRequest{
		Input: texts,
		Model: openai.EmbeddingModel(embeddingModel),
	}

	resp, err := e.retrier.Do(ctx, func() (openai.EmbeddingResponse, error) {
		return e.openaiClient.CreateEmbeddings(ctx, req)
	})
	if err != nil {
		logger.WithError(err).Error("failed to call openai to create embedding")
		return nil, err
	}

	result := make([]Embedding, 0)
	for _, data := range resp.Data {
		result = append(result, Embedding{
			Embedding: data.Embedding,
		})
	}

	return result, nil
}
