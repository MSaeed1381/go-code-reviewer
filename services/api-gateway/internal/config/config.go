package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Env        string        `yaml:"env" json:"env"`
	HttpServer HttpServer    `yaml:"http_server" json:"http_server"`
	Github     GithubSection `yaml:"github" json:"github"`
	Kafka      KafkaSection  `yaml:"kafka" json:"kafka"`
}

type GithubSection struct {
	WebhookSecret string `yaml:"webhook_secret" json:"webhook_secret"`
}

type HttpServer struct {
	Address string `yaml:"address" json:"address"`
}

type KafkaSection struct {
	Brokers string `yaml:"brokers" json:"brokers"`
	Topic   string `yaml:"topic" json:"topic"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{
		Github: GithubSection{
			WebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
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
