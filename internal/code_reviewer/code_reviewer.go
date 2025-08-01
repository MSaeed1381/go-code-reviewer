package code_reviewer

import (
	"context"
	"github.com/google/go-github/v58/github"
	"go_code_reviewer/internal/assistant"
	"go_code_reviewer/internal/embedder"
	"go_code_reviewer/internal/errors"
	"go_code_reviewer/internal/parser"
	"go_code_reviewer/pkg/log"
	"strings"
)

type Module struct {
	projectParser         *parser.ProjectParser
	projectEmbedder       *embedder.ProjectEmbedder
	codeAssistant         *assistant.Assistant
	pullRequestEventQueue chan *github.PullRequestEvent
	githubClient          *github.Client
}

func NewModule(projectParser *parser.ProjectParser, projectEmbedder *embedder.ProjectEmbedder, codeAssistant *assistant.Assistant, pullRequestEventQueue chan *github.PullRequestEvent, githubClient *github.Client) *Module {
	return &Module{
		projectParser:         projectParser,
		projectEmbedder:       projectEmbedder,
		codeAssistant:         codeAssistant,
		pullRequestEventQueue: pullRequestEventQueue,
		githubClient:          githubClient,
	}
}

func (m *Module) Start() {
	logger := log.GetLogger()
	ctx := context.Background()

	for {
		select {
		case prEvent := <-m.pullRequestEventQueue:
			review, err := m.reviewCode(ctx, prEvent)
			if err != nil {
				logger.WithError(err).Error("review code error")
				return
			}

			parts := strings.Split(prEvent.GetRepo().GetFullName(), "/")
			if len(parts) != 2 {
				logger.Error("invalid full name")
				return
			}

			owner, repo := parts[0], parts[1]
			if err = m.postPRComment(ctx, prEvent.GetNumber(), review, owner, repo); err != nil {
				logger.WithError(err).Error("post pr comment error")
				return
			}
		}
	}
}

func (m *Module) reviewCode(ctx context.Context, event *github.PullRequestEvent) (string, error) {
	logger := log.GetLogger()
	project, err := CloneProject(event.GetRepo().GetCloneURL(), event.GetPullRequest().GetHead().GetRef())
	if err != nil {
		logger.WithError(err).Error("failed to clone project")
		return "", err
	}

	snippets, err := m.projectParser.ParseProject(ctx, project.RepoPath)
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

	diff, err := downloadUrl(event.GetPullRequest().GetDiffURL())
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

func (m *Module) postPRComment(ctx context.Context, prNumber int, body, owner, repo string) error {
	comment := &github.IssueComment{Body: &body}
	_, _, err := m.githubClient.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	if err != nil {
		return err
	}
	return nil
}
