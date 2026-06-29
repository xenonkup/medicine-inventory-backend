package main

import (
	"context"
	"log"
	"strconv"

	"pharmacy-backend/internal/config"
	"pharmacy-backend/internal/database"
	"pharmacy-backend/internal/domain"
	"pharmacy-backend/internal/handler"
	"pharmacy-backend/internal/repository"
	"pharmacy-backend/internal/router"
	"pharmacy-backend/internal/service"
	"pharmacy-backend/pkg/jwt"
)

func main() {
	cfg := config.Load()

	// --- Database ---
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("startup: cannot connect to database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("startup: migration failed: %v", err)
	}

	// --- Dependency wiring (bottom-up) ---
	jwtMgr := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	txManager := repository.NewTxManager(db)
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	medicineRepo := repository.NewMedicineRepository(db)
	lotRepo := repository.NewLotRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	settingRepo := repository.NewSettingRepository(db)

	authService := service.NewAuthService(userRepo, jwtMgr)
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	medicineService := service.NewMedicineService(medicineRepo, categoryRepo, lotRepo)
	stockService := service.NewStockService(txManager, lotRepo, transactionRepo, medicineRepo)
	settingsService := service.NewSettingsService(settingRepo)
	dashboardService := service.NewDashboardService(medicineRepo, lotRepo, transactionRepo, settingsService, cfg.NearExpiryDays)
	reportService := service.NewReportService(transactionRepo, lotRepo)

	// Create the first admin if the database has no users yet.
	if err := userService.EnsureBootstrapAdmin(
		context.Background(),
		cfg.BootstrapAdminUsername,
		cfg.BootstrapAdminPassword,
		cfg.BootstrapAdminName,
	); err != nil {
		log.Printf("startup: bootstrap admin skipped: %v", err)
	}

	// Seed default settings (e.g. near-expiry window) if not present.
	if err := settingsService.EnsureDefault(
		context.Background(),
		domain.SettingNearExpiryDays,
		strconv.Itoa(cfg.NearExpiryDays),
	); err != nil {
		log.Printf("startup: seed settings skipped: %v", err)
	}

	handlers := router.Handlers{
		Auth:      handler.NewAuthHandler(authService),
		User:      handler.NewUserHandler(userService),
		Category:  handler.NewCategoryHandler(categoryService),
		Medicine:  handler.NewMedicineHandler(medicineService),
		Stock:     handler.NewStockHandler(stockService),
		Dashboard: handler.NewDashboardHandler(dashboardService),
		Report:    handler.NewReportHandler(reportService),
		Settings:  handler.NewSettingsHandler(settingsService),
	}

	r := router.New(cfg, jwtMgr, handlers)

	addr := ":" + cfg.Port
	log.Printf("startup: listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("startup: server stopped: %v", err)
	}
}
