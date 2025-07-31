package embedder

import (
	"context"
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/sashabaranov/go-openai"
	"go_code_reviewer/internal/parser"
	"go_code_reviewer/pkg/log"
)

type ProjectEmbedder struct {
	chromaCollection chroma.Collection
	embeddingClient  *openai.Client
	embeddingModel   openai.EmbeddingModel
}

func NewProjectEmbedder(embeddingClient *openai.Client, chromaCollection chroma.Collection, embeddingModel openai.EmbeddingModel) *ProjectEmbedder {
	return &ProjectEmbedder{
		chromaCollection: chromaCollection,
		embeddingClient:  embeddingClient,
		embeddingModel:   embeddingModel,
	}
}

func (p *ProjectEmbedder) EmbedProject(ctx context.Context, snippets []*parser.Snippet) error {
	logger := log.GetLogger()
	var texts []string
	for _, snippet := range snippets {
		texts = append(texts, snippet.Content)
	}

	req := openai.EmbeddingRequest{
		Input: texts,
		Model: p.embeddingModel,
	}
	resp, err := p.embeddingClient.CreateEmbeddings(ctx, req)
	if err != nil {
		logger.WithError(err).Error("Failed to create embedding")
		return err
	}

	for i := range snippets {
		snippets[i].Embedding = resp.Data[i].Embedding
	}
	return p.persistEmbeddings(snippets)
}

func (p *ProjectEmbedder) persistEmbeddings(snippets []*parser.Snippet) error {
	logger := log.GetLogger()
	var ids []chroma.DocumentID
	var documents []string
	var embeddingsList embeddings.Embeddings
	var metadataList []chroma.DocumentMetadata

	for _, snippet := range snippets {
		ids = append(ids, chroma.DocumentID(snippet.ID))
		documents = append(documents, snippet.Content)
		embeddingsList = append(embeddingsList, embeddings.NewEmbeddingFromFloat32(snippet.Embedding))
		metadataList = append(metadataList, chroma.NewMetadata(
			chroma.NewStringAttribute("filename", snippet.Filename),
			chroma.NewStringAttribute("language", string(snippet.Language)),
		))
	}

	err := p.chromaCollection.Add(
		context.Background(),
		chroma.WithIDs(ids...),
		chroma.WithEmbeddings(embeddingsList...),
		chroma.WithTexts(documents...),
		chroma.WithMetadatas(metadataList...),
	)
	if err != nil {
		logger.WithError(err).Error("Failed to persist embeddings")
		return err
	}

	return nil
}
