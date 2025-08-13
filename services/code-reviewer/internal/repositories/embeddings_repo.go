package repositories

import (
	"context"
	"fmt"
	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	"go_code_reviewer/services/code-reviewer/internal/models"
)

type EmbeddingsRepository interface {
	Add(ctx context.Context, snippets []*models.Snippet, projectId string) error
	GetNearestRecord(ctx context.Context, vectorEmbedding []float32, nResult int, projectId string) ([]*models.Snippet, error)
}

const projectIdKey = "project_id"

type EmbeddingRepositoryImpl struct {
	chromaCollection chroma.Collection
}

func NewEmbeddingRepository(chromaClient chroma.Client, embeddingFunction embeddings.EmbeddingFunction, collectionName string) EmbeddingsRepository {
	chromaCollection, err := chromaClient.GetOrCreateCollection(context.Background(), collectionName, chroma.WithEmbeddingFunctionCreate(embeddingFunction))
	if err != nil {
		panic(err)
	}

	return &EmbeddingRepositoryImpl{
		chromaCollection: chromaCollection,
	}
}

func (p *EmbeddingRepositoryImpl) Add(ctx context.Context, snippets []*models.Snippet, projectId string) error {
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
			chroma.NewStringAttribute("language", snippet.Language),
			chroma.NewStringAttribute(projectIdKey, projectId),
		))
	}

	return p.chromaCollection.Add(
		ctx,
		chroma.WithIDs(ids...),
		chroma.WithEmbeddings(embeddingsList...),
		chroma.WithTexts(documents...),
		chroma.WithMetadatas(metadataList...),
	)
}

func (p *EmbeddingRepositoryImpl) GetNearestRecord(ctx context.Context, vectorEmbedding []float32, nResult int, projectId string) ([]*models.Snippet, error) {
	var results, err = p.chromaCollection.Query(
		ctx,
		chroma.WithQueryEmbeddings(embeddings.NewEmbeddingFromFloat32(vectorEmbedding)),
		chroma.WithNResults(nResult),
		chroma.WithWhereQuery(chroma.EqString(projectIdKey, projectId)),
	)
	if err != nil {
		return nil, err
	}

	if len(results.GetDocumentsGroups()) == 0 {
		return []*models.Snippet{}, nil
	}
	documents := results.GetDocumentsGroups()[0]
	metadata := results.GetMetadatasGroups()[0]

	snippets := make([]*models.Snippet, 0)
	for i, doc := range documents {
		filename, _ := metadata[i].GetString("filename")
		language, _ := metadata[i].GetString("language")
		snippets = append(snippets, &models.Snippet{
			Content:  doc.ContentString(),
			Filename: filename,
			Language: language,
		})

		fmt.Println(doc.ContentString())
		fmt.Println("-------")
	}

	return snippets, nil
}
