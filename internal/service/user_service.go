package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/dto"
	"pharmacy-backend/internal/repository"
	"pharmacy-backend/pkg/hash"
)

// UserService handles user management (Admin).
type UserService struct {
	users repository.UserRepository
}

// NewUserService builds a UserService.
func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

// Create registers a new user.
func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest) (*domain.User, error) {
	hashed, err := hash.Password(req.Password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Username:     strings.TrimSpace(req.Username),
		PasswordHash: hashed,
		FullName:     strings.TrimSpace(req.FullName),
		Role:         domain.UserRole(req.Role),
		LineUserID:   req.LineUserID,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrDuplicateUsername
		}
		return nil, err
	}
	return user, nil
}

// List returns a page of users.
func (s *UserService) List(ctx context.Context, offset, limit int) ([]domain.User, int64, error) {
	return s.users.List(ctx, offset, limit)
}

// GetByID returns a single user.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}

// Update changes profile fields and role.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.FullName = strings.TrimSpace(req.FullName)
	user.Role = domain.UserRole(req.Role)
	user.LineUserID = req.LineUserID
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// SetStatus activates or deactivates a user.
func (s *UserService) SetStatus(ctx context.Context, id uuid.UUID, active bool) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.IsActive = active
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// ResetPassword sets a new password.
func (s *UserService) ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	hashed, err := hash.Password(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hashed
	return s.users.Update(ctx, user)
}

// EnsureBootstrapAdmin creates the first admin account if the table is empty.
func (s *UserService) EnsureBootstrapAdmin(ctx context.Context, username, password, fullName string) error {
	count, err := s.users.CountAll(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hashed, err := hash.Password(password)
	if err != nil {
		return err
	}
	admin := &domain.User{
		Username:     username,
		PasswordHash: hashed,
		FullName:     fullName,
		Role:         domain.RoleAdmin,
		IsActive:     true,
	}
	return s.users.Create(ctx, admin)
}

// isUniqueViolation reports whether err is a unique-constraint failure.
func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
