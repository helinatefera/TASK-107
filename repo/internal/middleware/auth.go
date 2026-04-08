package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/repo"
	"github.com/chargeops/api/internal/service"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

var skipPaths = map[string]bool{
	"/api/v1/auth/register":      true,
	"/api/v1/auth/login":         true,
	"/api/v1/auth/recover":       true,
	"/api/v1/auth/recover/reset": true,
	"/health":                    true,
}

// guestPaths allow unauthenticated access with role "guest".
var guestPaths = map[string]bool{
	"/api/v1/content/carousel":  true,
	"/api/v1/content/campaigns": true,
	"/api/v1/content/rankings":  true,
}

func Auth(db *sqlx.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if skipPaths[path] {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				// Allow guest access to content list endpoints
				if c.Request().Method == "GET" && guestPaths[path] {
					c.Set("role", "guest")
					c.Set("permissions", map[string]bool{"content.read": true})
					return next(c)
				}
				return apperror.ErrUnauthorized
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")

			hash := sha256.Sum256([]byte(token))
			tokenHash := hex.EncodeToString(hash[:])

			ctx := c.Request().Context()

			session, err := repo.GetSessionByTokenHash(ctx, db, tokenHash)
			if err != nil {
				return apperror.ErrUnauthorized
			}

			now := time.Now()

			// Check idle expiry - refresh endpoint handles its own idle check in the service
			if path != "/api/v1/auth/refresh" {
				if now.After(session.IdleExpiresAt) {
					return apperror.ErrSessionExpired
				}
			}

			// Check absolute expiry
			if now.After(session.AbsExpiresAt) {
				return apperror.ErrSessionExpired
			}

			// Verify device ID
			deviceID := c.Request().Header.Get("X-Device-Id")
			if deviceID != session.DeviceID {
				return apperror.ErrDeviceMismatch
			}

			// Load user
			user, err := repo.GetUserByID(ctx, db, session.UserID)
			if err != nil {
				return apperror.ErrUnauthorized
			}

			c.Set("user_id", user.ID)
			c.Set("role", user.Role)
			c.Set("org_id", user.OrgID)
			c.Set("session_id", session.ID)
			c.Set("session", session)
			c.Set("user", user)

			// Load effective permissions into context for RequirePermission middleware
			perms, err := service.GetEffectivePermissions(ctx, db, user.ID, user.Role)
			if err == nil {
				permSet := make(map[string]bool, len(perms))
				for _, p := range perms {
					if p.Granted {
						permSet[p.Name] = true
					}
				}
				c.Set("permissions", permSet)
			}

			return next(c)
		}
	}
}
