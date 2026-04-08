package unit_tests

import (
	"os"
	"testing"
	"time"

	"github.com/chargeops/api/internal/config"
)

func TestDefaultConfigValues(t *testing.T) {
	// Clear any env vars that might interfere
	envVars := []string{"PORT", "DATABASE_URL", "BCRYPT_COST", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "MIGRATIONS_DIR"}
	saved := make(map[string]string)
	for _, key := range envVars {
		saved[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range saved {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Fatalf("default Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.DatabaseURL != "postgres://chargeops:chargeops@localhost:5433/chargeops?sslmode=disable" {
		t.Fatalf("default DatabaseURL = %q, want default postgres URL", cfg.DatabaseURL)
	}
	if cfg.BcryptCost != 12 {
		t.Fatalf("default BcryptCost = %d, want 12", cfg.BcryptCost)
	}
	if cfg.SessionIdleTime != 30*time.Minute {
		t.Fatalf("default SessionIdleTime = %v, want 30m", cfg.SessionIdleTime)
	}
	if cfg.SessionAbsTime != 7*24*time.Hour {
		t.Fatalf("default SessionAbsTime = %v, want 7 days", cfg.SessionAbsTime)
	}
	if cfg.MaxOpenConns != 25 {
		t.Fatalf("default MaxOpenConns = %d, want 25", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Fatalf("default MaxIdleConns = %d, want 10", cfg.MaxIdleConns)
	}
	if cfg.MigrationsDir != "migrations" {
		t.Fatalf("default MigrationsDir = %q, want %q", cfg.MigrationsDir, "migrations")
	}
}

func TestLockoutParameters(t *testing.T) {
	cfg := config.Load()

	if cfg.LockoutWindow != 10*time.Minute {
		t.Fatalf("LockoutWindow = %v, want 10m", cfg.LockoutWindow)
	}
	if cfg.LockoutMax != 10 {
		t.Fatalf("LockoutMax = %d, want 10", cfg.LockoutMax)
	}
	if cfg.LockoutDuration != 15*time.Minute {
		t.Fatalf("LockoutDuration = %v, want 15m", cfg.LockoutDuration)
	}
}

func TestRecoveryTTL(t *testing.T) {
	cfg := config.Load()

	if cfg.RecoveryTTL != 1*time.Hour {
		t.Fatalf("RecoveryTTL = %v, want 1h", cfg.RecoveryTTL)
	}
}

func TestEnvOverridePort(t *testing.T) {
	orig := os.Getenv("PORT")
	defer func() {
		if orig != "" {
			os.Setenv("PORT", orig)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	os.Setenv("PORT", "9090")
	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Fatalf("PORT override: got %q, want %q", cfg.Port, "9090")
	}
}

func TestEnvOverrideDatabaseURL(t *testing.T) {
	orig := os.Getenv("DATABASE_URL")
	defer func() {
		if orig != "" {
			os.Setenv("DATABASE_URL", orig)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	os.Setenv("DATABASE_URL", "postgres://user:pass@remotehost:5432/mydb?sslmode=require")
	cfg := config.Load()

	if cfg.DatabaseURL != "postgres://user:pass@remotehost:5432/mydb?sslmode=require" {
		t.Fatalf("DATABASE_URL override: got %q, want custom URL", cfg.DatabaseURL)
	}
}
