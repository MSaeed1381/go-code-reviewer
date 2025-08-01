package code_reviewer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func downloadUrl(diffURL string) (string, error) {
	req, err := http.NewRequest("GET", diffURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	resp, err := http.DefaultClient.Do(req)
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

type CloneResult struct {
	RepoPath string
	Cleanup  func() error
}

func CloneProject(cloneUrl, branch string) (*CloneResult, error) {
	dir, err := os.MkdirTemp("", "gh-pr-*")
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
		return os.RemoveAll(dir)
	}

	cmd := exec.Command("git", "clone", "--depth=1", "--branch", branch, cloneUrl, dir)
	_, err = cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return nil, err
	}

	return &CloneResult{
		RepoPath: dir,
		Cleanup:  cleanup,
	}, nil
}
