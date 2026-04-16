package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	BcryptCost      int
	SessionIdleTime time.Duration
	SessionAbsTime  time.Duration
	LockoutWindow   time.Duration
	LockoutMax      int
	LockoutDuration time.Duration
	RecoveryTTL     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	MigrationsDir   string
	SeedAdminEmail  string
	SeedAdminPass   string
}

func Load() *Config {
	return &Config{
		Port:            envOrDefault("PORT", "8080"),
		DatabaseURL:     envOrDefault("DATABASE_URL", "postgres://chargeops:chargeops@localhost:5433/chargeops?sslmode=disable"),
		BcryptCost:      envIntOrDefault("BCRYPT_COST", 12),
		SessionIdleTime: 30 * time.Minute,
		SessionAbsTime:  7 * 24 * time.Hour,
		LockoutWindow:   10 * time.Minute,
		LockoutMax:      10,
		LockoutDuration: 15 * time.Minute,
		RecoveryTTL:     1 * time.Hour,
		MaxOpenConns:    envIntOrDefault("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    envIntOrDefault("DB_MAX_IDLE_CONNS", 10),
		MigrationsDir:   envOrDefault("MIGRATIONS_DIR", "migrations"),
		SeedAdminEmail:  os.Getenv("SEED_ADMIN_EMAIL"),
		SeedAdminPass:   os.Getenv("SEED_ADMIN_PASSWORD"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
