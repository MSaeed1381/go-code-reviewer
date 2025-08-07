package models

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
