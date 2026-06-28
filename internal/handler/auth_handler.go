package handler

import (
	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/middleware"
	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Login godoc
// @Summary  Login
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body      dto.LoginRequest  true  "credentials"
// @Success  200   {object}  response.Envelope
// @Router   /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	res, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// Refresh godoc
// @Summary  Refresh access token
// @Tags     Auth
// @Router   /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	res, err := h.auth.Refresh(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// Me godoc
// @Summary  Current user profile
// @Tags     Auth
// @Security BearerAuth
// @Router   /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	profile, err := h.auth.Me(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, profile)
}

// Logout is a client-side token discard for the stateless JWT scheme.
func (h *AuthHandler) Logout(c *gin.Context) {
	response.OK(c, gin.H{"message": "logged out"})
}
