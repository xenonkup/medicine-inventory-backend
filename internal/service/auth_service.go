package service

import (
	"context"

	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
	"pharmacy-backend/pkg/hash"
	"pharmacy-backend/pkg/jwt"
)

// AuthService handles authentication concerns.
type AuthService struct {
	users repository.UserRepository
	jwt   *jwt.Manager
}

// NewAuthService builds an AuthService.
func NewAuthService(users repository.UserRepository, jwtMgr *jwt.Manager) *AuthService {
	return &AuthService{users: users, jwt: jwtMgr}
}

// Login verifies credentials and issues access + refresh tokens.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		// Do not leak whether the username exists.
		return nil, domain.ErrInvalidCredentials
	}
	if !hash.Check(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}
	return s.issueTokens(user)
}

// Refresh validates a refresh token and issues a fresh token pair.
func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	claims, err := s.jwt.Parse(req.RefreshToken)
	if err != nil || claims.Type != jwt.RefreshToken {
		return nil, domain.ErrInvalidToken
	}
	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}
	return s.issueTokens(user)
}

// Me returns the current user's profile.
func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*dto.UserProfile, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := dto.NewUserProfile(user)
	return &profile, nil
}

func (s *AuthService) issueTokens(user *domain.User) (*dto.AuthResponse, error) {
	access, err := s.jwt.GenerateAccess(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, err
	}
	refresh, err := s.jwt.GenerateRefresh(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    s.jwt.AccessTTLSeconds(),
		User:         dto.NewUserProfile(user),
	}, nil
}
