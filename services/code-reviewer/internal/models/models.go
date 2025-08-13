package models

type Snippet struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Filename  string    `json:"filename"`
	Language  string    `json:"language"`
	ProjectId string    `json:"project_id"`
	Embedding []float32 `json:"embedding,omitempty"`
}

func NewSnippet(id, content, filename, language string) *Snippet {
	return &Snippet{
		ID:       id,
		Content:  content,
		Filename: filename,
		Language: language,
	}
}
