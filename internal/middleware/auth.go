package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/pkg/jwt"
	"pharmacy-backend/pkg/response"
)

// Context keys for values stashed by the auth middleware.
const (
	ctxUserID   = "userID"
	ctxUsername = "username"
	ctxRole     = "role"
)

// Auth verifies the Bearer access token and stashes the caller's identity.
func Auth(jwtMgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Abort(c, domain.ErrUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil || claims.Type != jwt.AccessToken {
			response.Abort(c, domain.ErrInvalidToken)
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// UserID returns the authenticated user's id from the context.
func UserID(c *gin.Context) uuid.UUID {
	if v, ok := c.Get(ctxUserID); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// Role returns the authenticated user's role from the context.
func Role(c *gin.Context) string {
	if v, ok := c.Get(ctxRole); ok {
		if r, ok := v.(string); ok {
			return r
		}
	}
	return ""
}
