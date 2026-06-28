package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/response"
)

// UserHandler exposes user-management endpoints (Admin only).
type UserHandler struct {
	users *service.UserService
}

// NewUserHandler builds a UserHandler.
func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// List godoc
// @Summary  List users
// @Tags     Users
// @Security BearerAuth
// @Router   /users [get]
func (h *UserHandler) List(c *gin.Context) {
	page, pageSize := paginationParams(c)
	offset := (page - 1) * pageSize

	users, total, err := h.users.List(c.Request.Context(), offset, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	profiles := make([]dto.UserProfile, 0, len(users))
	for i := range users {
		profiles = append(profiles, dto.NewUserProfile(&users[i]))
	}
	response.Paginated(c, profiles, page, pageSize, total)
}

// Create godoc
// @Summary  Create user
// @Tags     Users
// @Security BearerAuth
// @Router   /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	user, err := h.users.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.NewUserProfile(user))
}

// Get godoc
// @Summary  Get user by id
// @Tags     Users
// @Security BearerAuth
// @Router   /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewUserProfile(user))
}

// Update godoc
// @Summary  Update user
// @Tags     Users
// @Security BearerAuth
// @Router   /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	user, err := h.users.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewUserProfile(user))
}

// SetStatus godoc
// @Summary  Activate/deactivate user
// @Tags     Users
// @Security BearerAuth
// @Router   /users/{id}/status [patch]
func (h *UserHandler) SetStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	user, err := h.users.SetStatus(c.Request.Context(), id, *req.IsActive)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewUserProfile(user))
}

// ResetPassword godoc
// @Summary  Reset user password
// @Tags     Users
// @Security BearerAuth
// @Router   /users/{id}/password [patch]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}
	if err := h.users.ResetPassword(c.Request.Context(), id, req.Password); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "password reset"})
}

// --- helpers ---

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "invalid id", nil)
		return uuid.Nil, false
	}
	return id, true
}

// parseUUID parses a raw string into a UUID without writing a response.
func parseUUID(raw string) (uuid.UUID, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func paginationParams(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
