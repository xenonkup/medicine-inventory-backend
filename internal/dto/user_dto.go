package dto

// CreateUserRequest is the body for POST /users (Admin).
type CreateUserRequest struct {
	Username   string  `json:"username" binding:"required,min=3,max=50"`
	Password   string  `json:"password" binding:"required,min=6"`
	FullName   string  `json:"full_name" binding:"required"`
	Role       string  `json:"role" binding:"required,oneof=ADMIN STAFF"`
	LineUserID *string `json:"line_user_id"`
}

// UpdateUserRequest is the body for PUT /users/{id} (Admin).
type UpdateUserRequest struct {
	FullName   string  `json:"full_name" binding:"required"`
	Role       string  `json:"role" binding:"required,oneof=ADMIN STAFF"`
	LineUserID *string `json:"line_user_id"`
}

// UpdateStatusRequest toggles a user's active flag.
type UpdateStatusRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}

// ResetPasswordRequest sets a new password for a user.
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}
