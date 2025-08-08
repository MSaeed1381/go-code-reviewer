package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Env         string           `yaml:"env" json:"env"`
	WorkerCount int32            `yaml:"worker_count" json:"worker_count"`
	Prometheus  PrometheusConfig `yaml:"prometheus" json:"prometheus"`
	LLM         LLMSection       `yaml:"llm" json:"llm"`
	Embedding   EmbeddingSection `yaml:"embedding" json:"embedding"`
	Tasks       TasksSection     `yaml:"tasks" json:"tasks"`
	ChromaDB    ChromaDBSection  `yaml:"chroma_db" json:"chroma_db"`
	Github      GithubSection    `yaml:"github" json:"github"`
	Kafka       KafkaSection     `yaml:"kafka" json:"kafka"`
}

type PrometheusConfig struct {
	Address string `yaml:"address" json:"address"`
}

type KafkaSection struct {
	Brokers    string `yaml:"brokers" json:"brokers"`
	GroupID    string `yaml:"group_id" json:"group_id"`
	Topics     string `yaml:"topics" json:"topics"`
	AutoOffset string `yaml:"auto_offset" json:"auto_offset"`
}

type GithubSection struct {
	AccessToken string `yaml:"access_token" json:"access_token"`
}

type ChromaDBSection struct {
	Address        string `yaml:"address"`
	CollectionName string `yaml:"collection_name" json:"collection_name"`
}

type LLMSection struct {
	APIBaseURL  string  `yaml:"api_base_url"`
	OpenApiKey  string  `yaml:"openapi_key"`
	Model       string  `yaml:"model"`
	Temperature float32 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

type EmbeddingSection struct {
	APIBaseURL string `yaml:"api_base_url"`
	Model      string `yaml:"model"`
}

type TasksSection struct {
	DetectLanguage DetectLanguage `yaml:"detect_language"`
	CodeReview     TaskConfig     `yaml:"code_review"`
	CodeCompletion TaskConfig     `yaml:"code_completion"`
	CodeGeneration TaskConfig     `yaml:"code_generation"`
}

type Model struct {
	Name        string  `yaml:"name"`
	Temperature float32 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	Prefix      string  `yaml:"prefix"`
}

type TaskConfig struct {
	Model   Model         `yaml:"model"`
	Prompts PromptSection `yaml:"prompts"`
}

type PromptSection struct {
	ZeroShot string `yaml:"zero_shot"`
}

type DetectLanguage struct {
	Contextual      string `yaml:"contextual"`
	NaturalLanguage string `yaml:"natural_language"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{
		LLM: LLMSection{
			OpenApiKey: os.Getenv("LLM_OPEN_AI_API_KEY"),
		},
		Github: GithubSection{
			AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
		},
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
