package middleware

import (
	"github.com/chargeops/api/internal/apperror"
	"github.com/labstack/echo/v4"
)

func RequireRole(roles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(string)
			if !ok || !allowed[role] {
				return apperror.ErrForbidden
			}
			return next(c)
		}
	}
}

// RequirePermission checks that the authenticated user has ALL of the named
// permissions granted (via role defaults + user-level overrides).
// Permissions are loaded into context by the Auth middleware.
// Administrators bypass permission checks if no permissions are configured.
func RequirePermission(names ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			perms, _ := c.Get("permissions").(map[string]bool)
			// If no permissions are loaded (e.g. no seed data), allow admins through
			if len(perms) == 0 {
				role, _ := c.Get("role").(string)
				if role == "administrator" {
					return next(c)
				}
			}
			for _, name := range names {
				if !perms[name] {
					return apperror.ErrForbidden
				}
			}
			return next(c)
		}
	}
}
