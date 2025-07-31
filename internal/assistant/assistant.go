package assistant

import (
	"context"
	"fmt"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	"go_code_reviewer/internal/config"
	"go_code_reviewer/internal/errors"
	"go_code_reviewer/pkg/log"
	"strings"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/sashabaranov/go-openai"
)

type Assistant struct {
	config           *config.Config
	chromaCollection chroma.Collection
	llmClient        *openai.Client
	embeddingClient  *openai.Client
}

func NewAssistant(config *config.Config, chromaCollection chroma.Collection, llmClient *openai.Client) *Assistant {
	return &Assistant{
		config:           config,
		chromaCollection: chromaCollection,
		llmClient:        llmClient,
		embeddingClient:  llmClient,
	}
}

func (a *Assistant) PerformTask(ctx context.Context, task Task, queryText string) (string, error) {
	logger := log.GetLogger()
	contextString, err := a.getContextFromChroma(ctx, queryText)
	if err != nil {
		logger.WithError(err).Error("failed to get context from chroma")
		return "", err
	}

	prompt, err := a.buildPrompt(task, queryText, contextString, "Go")
	if err != nil {
		logger.WithError(err).Error("failed to build prompt")
		return "", err
	}
	logger.WithField("prompt", prompt).Info("build prompt")

	response, err := a.queryLLM(ctx, prompt)
	if err != nil {
		logger.WithError(err).Error("failed to query LLM")
		return "", err
	}

	return response, nil
}

func (a *Assistant) getContextFromChroma(ctx context.Context, queryText string) (string, error) {
	logger := log.GetLogger()
	resp, err := a.embeddingClient.CreateEmbeddings(ctx, &openai.EmbeddingRequest{
		Input: []string{queryText},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		logger.WithError(err).Error("failed to create embeddings")
		return "", err
	}
	queryEmbedding := resp.Data[0].Embedding

	results, err := a.chromaCollection.Query(
		ctx,
		chroma.WithQueryEmbeddings(embeddings.NewEmbeddingFromFloat32(queryEmbedding)),
		chroma.WithNResults(5),
	)
	if err != nil {
		logger.WithError(err).Error("failed to query chroma")
		return "", err
	}

	metadata := results.GetMetadatasGroups()[0]
	documents := results.GetDocumentsGroups()[0]
	var contextBuilder strings.Builder
	for i, doc := range documents {
		filename, _ := metadata[i].GetString("filename")
		contextBuilder.WriteString(fmt.Sprintf("--- Context Snippet %d from file %s ---\n", i, filename))
		contextBuilder.WriteString(doc.ContentString())
		contextBuilder.WriteString("\n\n")
	}

	return contextBuilder.String(), nil
}

func (a *Assistant) buildPrompt(task Task, queryText, contextString, language string) (string, error) {
	var template string
	switch task {
	case TaskCodeReview:
		template = a.config.Tasks.CodeReview.Prompts.ZeroShot
	case TaskCodeCompletion:
		template = a.config.Tasks.CodeCompletion.Prompts.ZeroShot
	case TaskCodeGeneration:
		template = a.config.Tasks.CodeGeneration.Prompts.ZeroShot
	default:
		return "", errors.ErrUnknownIntent
	}

	r := strings.NewReplacer(
		"{text}", queryText,
		"{context}", contextString,
		"{language}", language,
	)
	finalPrompt := r.Replace(template)

	return finalPrompt, nil
}

func (a *Assistant) queryLLM(ctx context.Context, prompt string) (string, error) {
	logger := log.GetLogger()
	resp, err := a.llmClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       a.config.LLM.Model,
			Temperature: a.config.LLM.Temperature,
			MaxTokens:   a.config.LLM.MaxTokens,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
		},
	)
	if err != nil {
		logger.WithError(err).Error("failed to create chat completion")
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.ErrNoResponseChoice
	}

	return resp.Choices[0].Message.Content, nil
}
