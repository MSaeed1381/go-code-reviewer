package vsc

import (
	"context"
	"fmt"
	"github.com/google/go-github/v58/github"
	"go_code_reviewer/pkg/retry"
	"io"
	"net/http"
	"os"
	"os/exec"
)

type Github struct {
	githubClient *github.Client
	retrier      retry.Retrier[*http.Response]
}

type GithubOption func(github *Github)

func WithRetry(retrier retry.Retrier[*http.Response]) GithubOption {
	return func(github *Github) {
		github.retrier = retrier
	}
}

func NewGithub(githubClient *github.Client, opts ...GithubOption) VersionControlSystem {
	g := &Github{
		githubClient: githubClient,
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.retrier == nil {
		g.retrier = retry.New[*http.Response](retry.Options{MaxRetries: 1})
	}

	return g
}

func (g *Github) DownloadUrl(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	resp, err := g.retrier.Do(ctx, func() (*http.Response, error) {
		return http.DefaultClient.Do(req)
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected response: %d\nBody: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (g *Github) Clone(ctx context.Context, url, branch string) (string, func() error, error) {
	dir, err := os.MkdirTemp("", "gh-pr-*")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() error {
		return os.RemoveAll(dir)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", "--branch", branch, url, dir)
	_, err = cmd.CombinedOutput()
	if err != nil {
		_ = cleanup()
		return "", nil, err
	}

	return dir, cleanup, nil
}

func (g *Github) PostPRComment(ctx context.Context, prNumber int, body, owner, repo string) error {
	comment := &github.IssueComment{Body: &body}
	_, err := g.retrier.Do(ctx, func() (*http.Response, error) {
		_, _, err := g.githubClient.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
