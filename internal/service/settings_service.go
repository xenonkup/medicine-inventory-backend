package service

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"

	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/repository"
)

// SettingsService reads/writes key/value system settings.
type SettingsService struct {
	repo repository.SettingRepository
}

// NewSettingsService builds a SettingsService.
func NewSettingsService(repo repository.SettingRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// List returns all settings.
func (s *SettingsService) List(ctx context.Context) ([]domain.Setting, error) {
	return s.repo.List(ctx)
}

// GetInt returns the integer value of a setting, or fallback if missing/invalid.
func (s *SettingsService) GetInt(ctx context.Context, key string, fallback int) int {
	setting, err := s.repo.Get(ctx, key)
	if err != nil || setting == nil {
		return fallback
	}
	if n, perr := strconv.Atoi(setting.Value); perr == nil {
		return n
	}
	return fallback
}

// Set upserts a setting value.
func (s *SettingsService) Set(ctx context.Context, key, value string, by uuid.UUID) (*domain.Setting, error) {
	setting := &domain.Setting{
		Key:       key,
		Value:     value,
		UpdatedBy: &by,
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	return setting, nil
}

// EnsureDefault seeds a setting if it does not exist yet.
func (s *SettingsService) EnsureDefault(ctx context.Context, key, value string) error {
	existing, err := s.repo.Get(ctx, key)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	return s.repo.Upsert(ctx, &domain.Setting{Key: key, Value: value, UpdatedAt: time.Now()})
}
