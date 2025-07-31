package parser

type Language string

var (
	LanguagePython Language = "python"
	LanguageGo     Language = "go"
)

type Snippet struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Filename  string    `json:"filename"`
	Language  Language  `json:"language"`
	Embedding []float32 `json:"embedding,omitempty"`
}

func NewSnippet(id, content, filename string, language Language) *Snippet {
	return &Snippet{
		ID:       id,
		Content:  content,
		Filename: filename,
		Language: language,
	}
}
