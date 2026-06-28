package middleware

import (
	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/pkg/response"
)

// RequireRole restricts a route to the given roles. It must run after Auth.
func RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[string(r)] = struct{}{}
	}
	return func(c *gin.Context) {
		role := Role(c)
		if _, ok := allowed[role]; !ok {
			response.Abort(c, domain.ErrForbidden)
			return
		}
		c.Next()
	}
}

// AdminOnly is a convenience for RequireRole(RoleAdmin).
func AdminOnly() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin)
}
