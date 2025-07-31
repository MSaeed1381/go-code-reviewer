package api

import (
	"github.com/gin-gonic/gin"
	"go_code_reviewer/internal/modules"
	"net/http"
)

type Handler struct {
	module *modules.Module
}

func NewHandler(module *modules.Module) *Handler {
	return &Handler{
		module: module,
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
