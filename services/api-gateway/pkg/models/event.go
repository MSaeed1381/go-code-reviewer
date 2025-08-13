package models

import "fmt"

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

func GetProjectIdentifier(pr *PullRequestEvent) string {
	return fmt.Sprintf("%s/%s/%s/%d", pr.Owner, pr.Repo, pr.Branch, pr.Number)
}
