package main

import (
	"context"
	"log"

	"pharmacy-backend/internal/config"
	"pharmacy-backend/internal/database"
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

	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	medicineRepo := repository.NewMedicineRepository(db)

	authService := service.NewAuthService(userRepo, jwtMgr)
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	medicineService := service.NewMedicineService(medicineRepo, categoryRepo)

	// Create the first admin if the database has no users yet.
	if err := userService.EnsureBootstrapAdmin(
		context.Background(),
		cfg.BootstrapAdminUsername,
		cfg.BootstrapAdminPassword,
		cfg.BootstrapAdminName,
	); err != nil {
		log.Printf("startup: bootstrap admin skipped: %v", err)
	}

	handlers := router.Handlers{
		Auth:     handler.NewAuthHandler(authService),
		User:     handler.NewUserHandler(userService),
		Category: handler.NewCategoryHandler(categoryService),
		Medicine: handler.NewMedicineHandler(medicineService),
	}

	r := router.New(cfg, jwtMgr, handlers)

	addr := ":" + cfg.Port
	log.Printf("startup: listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("startup: server stopped: %v", err)
	}
}
