package assistant

import (
	"context"
	"fmt"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	"go_code_reviewer/internal/config"
	log2 "go_code_reviewer/pkg/log"
	"log"
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

func (a *Assistant) PerformCodingTask(ctx context.Context, intent, queryText string) (string, error) {
	log.Printf("Performing task for intent: '%s'", intent)

	contextString, err := a.getContextFromChroma(ctx, queryText)
	if err != nil {
		return "", fmt.Errorf("failed to get context from chroma: %w", err)
	}
	if contextString == "" {
		log.Println("No relevant context found in ChromaDB.")
	} else {
		log.Println("Found relevant context from ChromaDB.")
		log.Println("context string ---------------")
		log.Println(contextString)
		log.Println("---------------")
	}

	prompt, err := a.buildPrompt(intent, queryText, contextString, "Go")
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	log.Printf("Final prompt being sent to LLM:\n---\n%s\n---", prompt) // برای دیباگ

	response, err := a.queryLLM(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to query LLM: %w", err)
	}

	return response, nil
}

func (a *Assistant) getContextFromChroma(ctx context.Context, queryText string) (string, error) {
	logger := log2.GetLogger()
	req := openai.EmbeddingRequest{
		Input: []string{queryText},
		Model: openai.SmallEmbedding3,
	}
	resp, err := a.embeddingClient.CreateEmbeddings(ctx, req)
	if err != nil {
		logger.WithError(err).Error("Failed to create embeddings")
		return "", err
	}
	queryEmbedding := resp.Data[0].Embedding

	results, err := a.chromaCollection.Query(
		ctx,
		chroma.WithQueryEmbeddings(embeddings.NewEmbeddingFromFloat32(queryEmbedding)),
		chroma.WithNResults(2),
	)
	if err != nil {
		logger.WithError(err).Error("Failed to query chroma")
		return "", err
	}

	log.Println("1111111")
	var contextBuilder strings.Builder
	for _, docGroups := range results.GetDocumentsGroups() {
		for _, doc := range docGroups {
			log.Println(doc.ContentString())
			log.Println("----")
			contextBuilder.WriteString(fmt.Sprintf("--- Context Snippet %d from file %s ---\n"))
			contextBuilder.WriteString(doc.ContentString())
			contextBuilder.WriteString("\n\n")
		}
	}

	return contextBuilder.String(), nil
}

func (a *Assistant) buildPrompt(intent, queryText, contextString, language string) (string, error) {
	var template string
	switch intent {
	case "code_review":
		template = a.config.Tasks.CodeReview.Prompts.ZeroShot
	case "code_completion":
		template = a.config.Tasks.CodeCompletion.Prompts.ZeroShot
	case "code_generation":
		template = a.config.Tasks.CodeGeneration.Prompts.ZeroShot
	default:
		return "", fmt.Errorf("unknown intent: %s", intent)
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
	log.Println("Sending prompt to LLM...")
	resp, err := a.llmClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       a.config.LLM.Model,
			Temperature: a.config.LLM.Temperature,
			MaxTokens:   a.config.LLM.MaxTokens,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}
