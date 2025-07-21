package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func (h *Handler) newErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	h.log.Error(message, slog.String("error", err.Error()))
	c.AbortWithStatusJSON(statusCode, ErrorResponse{Message: message})
}
