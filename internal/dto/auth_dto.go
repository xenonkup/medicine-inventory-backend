package dto

import "pharmacy-backend/internal/domain"

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserProfile is the safe, public view of a user.
type UserProfile struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	FullName string          `json:"full_name"`
	Role     domain.UserRole `json:"role"`
	IsActive bool            `json:"is_active"`
}

// AuthResponse is returned by login and refresh.
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         UserProfile `json:"user"`
}

// NewUserProfile maps a domain user to its public profile.
func NewUserProfile(u *domain.User) UserProfile {
	return UserProfile{
		ID:       u.ID.String(),
		Username: u.Username,
		FullName: u.FullName,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}
