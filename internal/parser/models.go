package parser

type Snippet struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Filename  string    `json:"filename"`
	Language  string    `json:"language"`
	Embedding []float32 `json:"embedding,omitempty"`
}

type Language string

var (
	LanguagePython Language = "python"
	LanguageGo     Language = "go"
)
