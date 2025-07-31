package code_reviewer

import (
	"context"
	"go_code_reviewer/internal/assistant"
	"go_code_reviewer/internal/embedder"
	"go_code_reviewer/internal/errors"
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

func (m *Module) ReviewCode(ctx context.Context, query string) (string, error) {
	logger := log.GetLogger()
	snippets, err := m.projectParser.ParseProject(ctx, "./Code-Review-Demo") // TODO: download project
	if err != nil {
		logger.WithError(err).Error("failed to parse project")
		return "", err
	}

	if len(snippets) == 0 {
		logger.Error("no snippets found")
		return "", errors.ErrNoSnippetFound
	}

	err = m.projectEmbedder.EmbedProject(ctx, snippets)
	if err != nil {
		logger.WithError(err).Error("Failed to embed project")
		return "", err
	}

	finalResponse, err := m.codeAssistant.PerformTask(ctx, assistant.TaskCodeReview, query)
	if err != nil {
		logger.WithError(err).Error("failed to perform coding task")
		return "", err
	}

	return finalResponse, nil
}
