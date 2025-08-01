package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v58/github"
	"go_code_reviewer/internal/code_reviewer"
	"go_code_reviewer/internal/config"
	"net/http"
)

type Handler struct {
	config                *config.Config
	module                *code_reviewer.Module
	pullRequestEventQueue chan *github.PullRequestEvent
}

func NewHandler(config *config.Config, module *code_reviewer.Module, pullRequestEventQueue chan *github.PullRequestEvent) *Handler {
	return &Handler{
		config:                config,
		module:                module,
		pullRequestEventQueue: pullRequestEventQueue,
	}
}

func (h *Handler) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": 1})
}

func (h *Handler) readiness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": 1})
}

func (h *Handler) RegisterRoutes() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/readiness", h.readiness)
	r.POST("/liveness", h.liveness)
	r.POST("/webhook", h.webhook)

	return r
}
