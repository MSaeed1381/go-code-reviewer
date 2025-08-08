package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v58/github"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/api-gateway/pkg/models"
	"strings"
)

func (h *Handler) githubWebhook(c *gin.Context) {
	logger := log.GetLogger()
	logger.Info("Received request for github webhook")

	payload, err := github.ValidatePayload(c.Request, []byte(h.config.Github.WebhookSecret))
	if err != nil {
		h.handleErrorApiResponse(c, err, "failed to validate payload")
		return
	}

	githubEvent, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
	if err != nil {
		h.handleErrorApiResponse(c, err, "failed to parse webhook")
		return
	}

	switch e := githubEvent.(type) {
	case *github.PullRequestEvent:
		event := convertGitHubEvent(e)
		if event == nil {
			h.handleErrorApiResponse(c, err, "failed to convert github event")
			return
		}
		logger.Infof("Received pull request event %v", event)
		err = h.module.ProcessEvent(event)
		if err != nil {
			h.handleErrorApiResponse(c, err, "failed to send event to kafka")
			return
		}
		logger.Info("Successfully send webhook to kafka")
		h.handleSuccessfulApiResponse(c, "received pull request event")
		return
	}

	h.handleSuccessfulApiResponse(c, "event not found")
}

func convertGitHubEvent(event *github.PullRequestEvent) *models.PullRequestEvent {
	if event == nil || event.PullRequest == nil || event.Repo == nil || event.Repo.Owner == nil {
		return nil
	}

	parts := strings.Split(event.GetRepo().GetFullName(), "/")
	if len(parts) != 2 {
		return nil
	}

	return &models.PullRequestEvent{
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
