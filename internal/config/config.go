package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration loaded from the environment.
type Config struct {
	Port    string
	AppEnv  string
	IsProd  bool

	DatabaseURL string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	CORSAllowedOrigins []string

	BootstrapAdminUsername string
	BootstrapAdminPassword string
	BootstrapAdminName     string

	NearExpiryDays int

	LineChannelAccessToken string
}

// Load reads configuration from a .env file (if present) and the environment.
// Missing optional values fall back to sensible defaults.
func Load() *Config {
	// .env is optional; in production the platform injects real env vars.
	if err := godotenv.Load(); err != nil {
		log.Println("config: no .env file found, relying on environment variables")
	}

	appEnv := getEnv("APP_ENV", "development")

	cfg := &Config{
		Port:                   getEnv("PORT", "8080"),
		AppEnv:                 appEnv,
		IsProd:                 appEnv == "production",
		DatabaseURL:            mustEnv("DATABASE_URL"),
		JWTSecret:              mustEnv("JWT_SECRET"),
		JWTAccessTTL:           getDuration("JWT_ACCESS_TTL", time.Hour),
		JWTRefreshTTL:          getDuration("JWT_REFRESH_TTL", 168*time.Hour),
		CORSAllowedOrigins:     getCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		BootstrapAdminUsername: getEnv("BOOTSTRAP_ADMIN_USERNAME", "admin"),
		BootstrapAdminPassword: getEnv("BOOTSTRAP_ADMIN_PASSWORD", "admin1234"),
		BootstrapAdminName:     getEnv("BOOTSTRAP_ADMIN_NAME", "System Administrator"),
		NearExpiryDays:         getInt("NEAR_EXPIRY_DAYS", 180),
		LineChannelAccessToken: getEnv("LINE_CHANNEL_ACCESS_TOKEN", ""),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		log.Fatalf("config: required environment variable %s is not set", key)
	}
	return v
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		log.Printf("config: invalid duration for %s=%q, using default", key, v)
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		log.Printf("config: invalid int for %s=%q, using default", key, v)
	}
	return fallback
}

func getCSV(key string, fallback []string) []string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return fallback
}
