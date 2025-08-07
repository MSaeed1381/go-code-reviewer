package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/api-gateway/internal/config"
	serviceErrors "go_code_reviewer/services/api-gateway/internal/errors"
	eventprocessor "go_code_reviewer/services/api-gateway/internal/event-processor"
	"net/http"
)

type Handler struct {
	config *config.Config
	module *eventprocessor.Module
}

func NewHandler(config *config.Config, module *eventprocessor.Module) *Handler {
	return &Handler{
		config: config,
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
	r.POST("/github-webhook", h.githubWebhook)

	return r
}

type SuccessResponse struct {
	Ok     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
}

func (h *Handler) handleSuccessfulApiResponse(c *gin.Context, response interface{}) {
	marshalled, err := json.Marshal(response)
	if err != nil {
		h.handleErrorApiResponse(c, err, "failed to marshal response")
	}

	resp := &SuccessResponse{
		Ok:     true,
		Result: marshalled,
	}
	c.JSON(http.StatusOK, resp)
	c.Abort()
}

type ErrorResponse struct {
	Ok        bool   `json:"ok"`
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

func (h *Handler) handleErrorApiResponse(c *gin.Context, err error, prompt string) {
	logger := log.GetLogger()

	var httpErr *serviceErrors.HttpError
	if errors.As(err, &httpErr) {
		if !httpErr.IsUserError {
			logger.WithError(err).Error(prompt)
		}
		resp := &ErrorResponse{
			Ok:        false,
			ErrorCode: httpErr.StatusCode,
			Message:   httpErr.Error(),
		}
		c.JSON(httpErr.StatusCode, resp)
		c.Abort()
		return
	}

	logger.WithError(err).Error(prompt)
	resp := &ErrorResponse{
		Ok:        false,
		ErrorCode: http.StatusInternalServerError,
		Message:   "Internal Server Error",
	}
	c.JSON(http.StatusInternalServerError, resp)
	c.Abort()
	return
}
