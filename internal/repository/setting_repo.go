package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pharmacy-backend/internal/domain"
)

type settingRepository struct {
	db *gorm.DB
}

// NewSettingRepository builds a GORM-backed SettingRepository.
func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) Get(ctx context.Context, key string) (*domain.Setting, error) {
	var s domain.Setting
	err := dbFromCtx(ctx, r.db).First(&s, "key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // missing setting is not an error; caller uses a default
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settingRepository) List(ctx context.Context) ([]domain.Setting, error) {
	var settings []domain.Setting
	err := dbFromCtx(ctx, r.db).Order("key ASC").Find(&settings).Error
	return settings, err
}

func (r *settingRepository) Upsert(ctx context.Context, setting *domain.Setting) error {
	return dbFromCtx(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_by", "updated_at"}),
		}).
		Create(setting).Error
}
