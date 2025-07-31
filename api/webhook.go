package api

import (
	"github.com/gin-gonic/gin"
)

type WebhookRequest struct {
	Query string `json:"query" form:"query" binding:"required"`
}

func (h *Handler) webhook(c *gin.Context) {
	req := WebhookRequest{}
	err := c.ShouldBind(&req)
	if err != nil {
		return
	}

	review, err := h.module.ReviewCode(c, req.Query)
	if err != nil {
		h.handleErrorApiResponse(c, err, "failed to review code")
		return
	}

	h.handleSuccessfulApiResponse(c, review)
}
