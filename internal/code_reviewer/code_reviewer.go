package code_reviewer

import (
	"context"
	"go_code_reviewer/internal/assistant"
	"go_code_reviewer/internal/embedder"
	"go_code_reviewer/internal/errors"
	"go_code_reviewer/internal/parser"
	"go_code_reviewer/internal/vsc"
	"go_code_reviewer/pkg/log"
)

type Module struct {
	projectParser         *parser.ProjectParser
	projectEmbedder       *embedder.ProjectEmbedder
	codeAssistant         *assistant.Assistant
	versionControl        vsc.VersionControlSystem
	pullRequestEventQueue chan *vsc.PullRequestEvent
}

func NewModule(projectParser *parser.ProjectParser, projectEmbedder *embedder.ProjectEmbedder, codeAssistant *assistant.Assistant, pullRequestEventQueue chan *vsc.PullRequestEvent, versionControl vsc.VersionControlSystem) *Module {
	return &Module{
		projectParser:         projectParser,
		projectEmbedder:       projectEmbedder,
		codeAssistant:         codeAssistant,
		pullRequestEventQueue: pullRequestEventQueue,
		versionControl:        versionControl,
	}
}

func (m *Module) Start() {
	logger := log.GetLogger()
	ctx := context.Background()

	for {
		select {
		case prEvent := <-m.pullRequestEventQueue:
			logger.WithField("number", prEvent.Number).Info("Received pull request event")
			review, err := m.reviewCode(ctx, prEvent)
			if err != nil {
				logger.WithError(err).Error("review code error")
				return
			}

			if err = m.versionControl.PostPRComment(ctx, prEvent.Number, review, prEvent.Owner, prEvent.Repo); err != nil {
				logger.WithError(err).Error("post pr comment error")
				return
			}
		}
	}
}

func (m *Module) reviewCode(ctx context.Context, event *vsc.PullRequestEvent) (string, error) {
	logger := log.GetLogger()
	repoPath, cleanup, err := m.versionControl.Clone(ctx, event.CloneURL, event.Branch)
	if err != nil {
		logger.WithError(err).Error("failed to clone project")
		return "", err
	}
	defer cleanup()

	snippets, err := m.projectParser.ParseProject(ctx, repoPath)
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

	diff, err := m.versionControl.DownloadUrl(ctx, event.DiffURL)
	if err != nil {
		logger.WithError(err).Error("failed to download url")
		return "", err
	}

	finalResponse, err := m.codeAssistant.PerformTask(ctx, assistant.TaskCodeReview, diff)
	if err != nil {
		logger.WithError(err).Error("failed to perform coding task")
		return "", err
	}

	return finalResponse, nil
}
