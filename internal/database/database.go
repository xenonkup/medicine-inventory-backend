package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"pharmacy-backend/internal/config"
	"pharmacy-backend/internal/domain"
)

// Connect opens a GORM connection to PostgreSQL.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	logLevel := gormlogger.Warn
	if !cfg.IsProd {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate runs AutoMigrate for the entities defined so far. Additional
// entities (Category, Medicine, Lot, StockTransaction, ...) are appended here
// as later phases land.
func Migrate(db *gorm.DB) error {
	log.Println("database: running auto-migration")
	return db.AutoMigrate(
		&domain.User{},
		&domain.Category{},
		&domain.Medicine{},
		&domain.Lot{},
		&domain.StockTransaction{},
		&domain.Setting{},
	)
}
