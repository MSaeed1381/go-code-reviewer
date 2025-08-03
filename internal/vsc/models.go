package vsc

import (
	"context"
	"github.com/google/go-github/v58/github"
	"strings"
)

type VersionControlSystem interface {
	DownloadUrl(ctx context.Context, url string) (string, error)
	Clone(ctx context.Context, url, branch string) (string, func() error, error)
	PostPRComment(ctx context.Context, prNumber int, body, owner, repo string) error
}

type PullRequestEvent struct {
	Owner    string
	Repo     string
	Number   int
	CloneURL string
	Branch   string
	Title    string
	Author   string
	DiffURL  string
}

func ConvertGitHubEvent(event *github.PullRequestEvent) *PullRequestEvent {
	if event == nil || event.PullRequest == nil || event.Repo == nil || event.Repo.Owner == nil {
		return nil
	}

	parts := strings.Split(event.GetRepo().GetFullName(), "/")
	if len(parts) != 2 {
		return nil
	}

	return &PullRequestEvent{
		Owner:    parts[0],
		Repo:     parts[1],
		Number:   event.GetPullRequest().GetNumber(),
		CloneURL: event.GetRepo().GetCloneURL(),
		Branch:   event.GetPullRequest().GetHead().GetRef(),
		Title:    event.GetPullRequest().GetTitle(),
		Author:   event.GetPullRequest().GetUser().GetLogin(),
		DiffURL:  event.GetPullRequest().GetDiffURL(),
	}
}
