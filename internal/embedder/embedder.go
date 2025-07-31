package embedder

import (
	"context"
	"fmt"
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/sashabaranov/go-openai"
	"go_code_reviewer/internal/config"
	"go_code_reviewer/internal/parser"
	"log"
)

type ProjectEmbedder struct {
	chromaCollection chroma.Collection
	embeddingClient  *openai.Client
	embeddingModel   openai.EmbeddingModel
	config           *config.Config
}

func NewProjectEmbedder(config *config.Config, embeddingClient *openai.Client) (*ProjectEmbedder, error) {
	chromaClient, err := chroma.NewHTTPClient(chroma.WithBaseURL(config.ChromaDB.Address))
	if err != nil {
		return nil, fmt.Errorf("failed to create chroma client: %w", err)
	}

	collectionMetadata := chroma.NewMetadata(
		chroma.NewStringAttribute("hnsw:space", "cosine"),
	)

	chromaCollection, err := chromaClient.GetOrCreateCollection(
		context.Background(),
		CollectionName,
		chroma.WithCollectionMetadataCreate(collectionMetadata),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create/get collection: %w", err)
	}
	log.Printf("ChromaDB collection '%s' is ready.", CollectionName)

	return &ProjectEmbedder{
		chromaCollection: chromaCollection,
		embeddingClient:  embeddingClient,
		embeddingModel:   openai.SmallEmbedding3,
		config:           config,
	}, nil
}

func (pe *ProjectEmbedder) EmbedProject(ctx context.Context, snippets []*parser.Snippet) error {
	log.Println("Starting to embed project snippets...")

	texts := make([]string, len(snippets))
	for i, s := range snippets {
		texts[i] = s.Content
	}

	req := openai.EmbeddingRequest{
		Input: texts,
		Model: pe.embeddingModel,
	}

	resp, err := pe.embeddingClient.CreateEmbeddings(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create embeddings: %w", err)
	}

	for i := range snippets {
		snippets[i].Embedding = resp.Data[i].Embedding
	}
	log.Printf("Successfully generated %d embeddings.", len(snippets))

	return pe.persistEmbeddings(snippets)
}

func (pe *ProjectEmbedder) persistEmbeddings(snippets []*parser.Snippet) error {
	log.Println("Storing embeddings in ChromaDB...")

	ids := make([]chroma.DocumentID, len(snippets))
	documents := make([]string, len(snippets))
	embeddingss := make([]embeddings.Embedding, len(snippets))
	metadatas := make([]chroma.DocumentMetadata, len(snippets))

	for i, s := range snippets {
		ids[i] = chroma.DocumentID(s.ID)
		documents[i] = s.Content
		embeddingss[i] = embeddings.NewEmbeddingFromFloat32(s.Embedding)
		metadatas[i] = chroma.NewMetadata(
			chroma.NewStringAttribute("filename", s.Filename),
			chroma.NewStringAttribute("language", s.Language),
		)
	}

	err := pe.chromaCollection.Add(
		context.Background(),
		chroma.WithIDs(ids...),
		chroma.WithEmbeddings(embeddingss...),
		chroma.WithTexts(documents...),
		chroma.WithMetadatas(metadatas...),
	)

	if err == nil {
		log.Println("Code ingestion and storage completed successfully!")
	}
	return err
}
