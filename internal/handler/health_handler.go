package handler

import (
	"github.com/gin-gonic/gin"
	"pharmacy-backend/pkg/response"
)

// Health is a liveness/readiness probe for Render.
func Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
