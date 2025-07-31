package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	serviceErrors "go_code_reviewer/internal/errors"
	"go_code_reviewer/pkg/log"
	"net/http"
)

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
