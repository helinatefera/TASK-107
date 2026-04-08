package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/repo"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func RateLimit(db *sqlx.DB, cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Read body and restore it
			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return apperror.ErrBadRequest
			}
			c.Request().Body = io.NopCloser(bytes.NewReader(body))

			var payload struct {
				Email string `json:"email"`
			}
			if err := json.Unmarshal(body, &payload); err != nil || payload.Email == "" {
				// Can't determine user, let handler deal with validation
				return next(c)
			}

			ctx := c.Request().Context()

			user, err := repo.GetUserByEmail(ctx, db, payload.Email)
			if err != nil {
				// User not found; let handler deal with it
				return next(c)
			}

			now := time.Now()

			// Check if already locked
			if user.LockedUntil.Valid && user.LockedUntil.Time.After(now) {
				return apperror.ErrAccountLocked
			}

			// Count failed attempts in lockout window
			since := now.Add(-cfg.LockoutWindow)
			count, err := repo.CountFailedAttempts(ctx, db, user.ID, since)
			if err != nil {
				return next(c)
			}

			if count >= cfg.LockoutMax {
				lockUntil := now.Add(cfg.LockoutDuration)
				_ = repo.SetLockedUntil(ctx, db, user.ID, lockUntil)
				return apperror.ErrAccountLocked
			}

			return next(c)
		}
	}
}
