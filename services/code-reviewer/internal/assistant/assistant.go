package assistant

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/code-reviewer/internal/config"
	"go_code_reviewer/services/code-reviewer/internal/embedder"
	"go_code_reviewer/services/code-reviewer/internal/repositories"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Assistant struct {
	config          *config.Config
	embeddingRepo   repositories.EmbeddingsRepository
	llm             llms.Model
	embeddingClient embedder.EmbeddingClient
}

func NewAssistant(config *config.Config, embeddingRepo repositories.EmbeddingsRepository, llm llms.Model, embeddingClient embedder.EmbeddingClient) *Assistant {
	return &Assistant{
		config:          config,
		embeddingRepo:   embeddingRepo,
		llm:             llm,
		embeddingClient: embeddingClient,
	}
}

func (a *Assistant) PerformTask(ctx context.Context, task Task, queryText, projectId string) (string, error) {
	logger := log.GetLogger()
	contextString, err := a.getContextFromChroma(ctx, projectId, queryText)
	if err != nil {
		logger.WithError(err).Error("failed to get context from chroma")
		return "", err
	}

	response, err := a.callLLMToPerformTask(ctx, task, queryText, contextString, "go")
	if err != nil {
		logger.WithError(err).Error("failed to query LLM")
		return "", err
	}

	return response, nil
}

func (a *Assistant) getContextFromChroma(ctx context.Context, projectId, queryText string) (string, error) {
	logger := log.GetLogger()

	resp, err := a.embeddingClient.CreateEmbeddings(ctx, string(openai.SmallEmbedding3), []string{queryText})
	if err != nil {
		logger.WithError(err).Error("failed to create embeddings")
		return "", err
	}

	records, err := a.embeddingRepo.GetNearestRecord(ctx, resp[0].Embedding, 5, projectId)
	if err != nil {
		logger.WithError(err).Error("failed to get nearest records")
		return "", err
	}

	var contextBuilder strings.Builder
	for i, record := range records {
		contextBuilder.WriteString(fmt.Sprintf("--- Context Snippet %d from file %s ---\n", i, record.Filename))
		contextBuilder.WriteString(record.Content)
		contextBuilder.WriteString("\n\n")
	}

	return contextBuilder.String(), nil
}

func (a *Assistant) callLLMToPerformTask(ctx context.Context, task Task, queryText, contextString, language string) (string, error) {
	logger := log.GetLogger()
	logger.WithFields(logrus.Fields{
		"query":   queryText,
		"context": contextString,
	}).Info("querying llm")

	var template string
	switch task {
	case TaskCodeReview:
		template = a.config.Tasks.CodeReview.Prompts.ZeroShot
	case TaskCodeCompletion:
		template = a.config.Tasks.CodeCompletion.Prompts.ZeroShot
	case TaskCodeGeneration:
		template = a.config.Tasks.CodeGeneration.Prompts.ZeroShot
	}

	promptTemplate := prompts.NewPromptTemplate(template, []string{"text", "context", "language"})
	chain := chains.NewLLMChain(a.llm, promptTemplate)

	prompt, err := chain.Prompt.FormatPrompt(map[string]any{
		"text":     queryText,
		"context":  contextString,
		"language": language,
	})
	if err != nil {
		logger.WithError(err).Error("failed to format prompt")
		return "", err
	}
	logger.WithField("prompt", prompt.String()).Info("prompt created")

	result, err := chains.Predict(ctx, chain, map[string]any{
		"text":     queryText,
		"context":  contextString,
		"language": language,
	}, chains.WithMaxTokens(a.config.LLM.MaxTokens))
	if err != nil {
		logger.WithError(err).Error("failed to call llm")
		return "", err
	}

	return result, nil
}
