package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	userCtxKey = contextKey("userID")
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			h.newErrorResponse(c, http.StatusUnauthorized, "authorization header is empty", fmt.Errorf("authorization header is empty"))
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			h.newErrorResponse(c, http.StatusUnauthorized, "invalid authorization header format", fmt.Errorf("invalid authorization header format"))
			return
		}

		tokenString := headerParts[1]
		claims, err := h.TokenManager.ParseToken(tokenString)
		if err != nil {
			h.newErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
			return
		}

		c.Set(string(userCtxKey), claims.UserID)
		c.Next()
	}
}

func GetUserIDFromCtx(c *gin.Context) (int64, bool) {
	val, ok := c.Get(string(userCtxKey))
	if !ok {
		return 0, false
	}

	userID, ok := val.(int64)
	return userID, ok
}
