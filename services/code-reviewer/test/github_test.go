package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v58/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go_code_reviewer/services/code-reviewer/internal/vsc"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadUrl_Success(t *testing.T) {
	expectedDiff := "diff --git a/file.txt b/file.txt\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-hello\n+world"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github.v3.diff", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedDiff)
	}))
	defer server.Close()

	g := vsc.NewGithub(nil)
	actualDiff, err := g.DownloadUrl(context.Background(), server.URL)
	require.NoError(t, err)
	assert.Equal(t, expectedDiff, actualDiff)
}

func TestDownloadUrl_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	}))
	defer server.Close()

	g := vsc.NewGithub(nil)
	_, err := g.DownloadUrl(context.Background(), server.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response: 404")
}

func TestClone_Success(t *testing.T) {
	repoURL := "https://github.com/git-fixtures/basic.git"
	branch := "master"

	g := vsc.NewGithub(nil)
	dir, cleanup, err := g.Clone(context.Background(), repoURL, branch)
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	_, err = os.Stat(dir)
	assert.False(t, os.IsNotExist(err), "directory should exist after clone")

	gitDirInfo, err := os.Stat(filepath.Join(dir, ".git"))
	assert.NoError(t, err, ".git directory should exist")
	assert.True(t, gitDirInfo.IsDir())

	err = cleanup()
	assert.NoError(t, err, "cleanup should not return an error")

	_, err = os.Stat(dir)
	assert.True(t, os.IsNotExist(err), "directory should not exist after cleanup")
}

func TestPostPRComment_Success(t *testing.T) {
	owner := "test-owner"
	repo := "test-repo"
	prNumber := 42
	expectedBody := "This is a test comment"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedURL := fmt.Sprintf("/api/v3/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)
		assert.Equal(t, expectedURL, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var comment github.IssueComment
		err = json.Unmarshal(bodyBytes, &comment)
		require.NoError(t, err)
		assert.Equal(t, expectedBody, *comment.Body)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id": 1, "body": "This is a test comment"}`)
	}))
	defer server.Close()

	testClient, err := github.NewClient(server.Client()).WithEnterpriseURLs(server.URL, server.URL)
	require.NoError(t, err)

	g := vsc.NewGithub(testClient)
	err = g.PostPRComment(context.Background(), prNumber, expectedBody, owner, repo)
	assert.NoError(t, err)
}
