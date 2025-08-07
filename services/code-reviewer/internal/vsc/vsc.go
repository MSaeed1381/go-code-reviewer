package vsc

import (
	"context"
)

type VersionControlSystem interface {
	DownloadUrl(ctx context.Context, url string) (string, error)
	Clone(ctx context.Context, url, branch string) (string, func() error, error)
	PostPRComment(ctx context.Context, prNumber int, body, owner, repo string) error
}
