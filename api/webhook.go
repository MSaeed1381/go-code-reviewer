package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v58/github"
	"go_code_reviewer/internal/vsc"
)

func (h *Handler) webhook(c *gin.Context) {
	payload, err := github.ValidatePayload(c.Request, []byte(h.config.Github.WebhookSecret))
	if err != nil {
		h.handleErrorApiResponse(c, err, "failed to validate payload")
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
	if err != nil {
		h.handleErrorApiResponse(c, err, "failed to parse webhook")
		return
	}

	switch e := event.(type) {
	case *github.PullRequestEvent:
		h.pullRequestEventQueue <- vsc.ConvertGitHubEvent(e)
		h.handleSuccessfulApiResponse(c, "received pull request event")
		return
	}

	h.handleSuccessfulApiResponse(c, "event not found")
}
