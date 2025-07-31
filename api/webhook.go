package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type WebhookRequest struct {
	Query  string `json:"query" form:"query"`
	Indent string `json:"indent" form:"indent"`
}

func (h *Handler) webhook(c *gin.Context) {
	req := WebhookRequest{}
	err := c.ShouldBind(&req)
	if err != nil {
		return
	}

	if req.Query == "" || req.Indent == "" {
		fmt.Println(req.Query)
		fmt.Println(req.Indent)
		c.JSON(http.StatusBadRequest, gin.H{"error": "query or body is empty"})
		return
	}

	review, err := h.module.ReviewCode(c, req.Query, req.Indent)
	if err != nil { // TODO: map error
		c.JSONP(http.StatusBadRequest, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": gin.H{
			"review": review,
		},
	})
}
