package modules

import (
	"context"
	"errors"
	"go_code_reviewer/internal/assistant"
	"go_code_reviewer/internal/embedder"
	"go_code_reviewer/internal/parser"
	"go_code_reviewer/pkg/log"
)

type Module struct {
	projectParser   *parser.ProjectParser
	projectEmbedder *embedder.ProjectEmbedder
	codeAssistant   *assistant.Assistant
}

func NewModule(projectParser *parser.ProjectParser, projectEmbedder *embedder.ProjectEmbedder, codeAssistant *assistant.Assistant) *Module {
	return &Module{
		projectParser:   projectParser,
		projectEmbedder: projectEmbedder,
		codeAssistant:   codeAssistant,
	}
}

func (m *Module) ReviewCode(ctx context.Context, query, indent string) (string, error) {
	logger := log.GetLogger()
	snippets, err := m.projectParser.ParseProject(ctx, "./Code-Review-Demo")
	if err != nil {
		logger.WithError(err).Error("Failed to parse project")
		return "", err
	}

	for _, snippet := range snippets {
		logger.Info("Snippet: %v", snippet)
		logger.Info("--------------------------")
	}

	if len(snippets) == 0 {
		logger.Error("No snippets found")
		return "", errors.New("no snippets found")
	}

	err = m.projectEmbedder.EmbedProject(ctx, snippets)
	if err != nil {
		logger.WithError(err).Error("Failed to embed project")
		return "", err
	}

	finalResponse, err := m.codeAssistant.PerformCodingTask(context.Background(), indent, query)
	if err != nil {
		logger.WithError(err).Error("Failed to perform coding task")
		return "", err
	}

	return finalResponse, nil
}
