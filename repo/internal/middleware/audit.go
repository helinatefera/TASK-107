package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// extractAuditEntity parses the URL path to find the primary resource type and ID.
// For nested/action routes like /api/v1/versions/:id/activate or /api/v1/users/:id/role,
// it finds the UUID and uses the segment before it as the entity type.
func extractAuditEntity(path string) (entityType, entityID string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// Walk backwards to find the first UUID-shaped segment (36 chars with hyphens)
	for i := len(parts) - 1; i >= 0; i-- {
		if len(parts[i]) == 36 && strings.Count(parts[i], "-") == 4 {
			entityID = parts[i]
			if i > 0 {
				entityType = parts[i-1]
			}
			return
		}
	}
	// No UUID found — use last segment as entity type (e.g., POST /api/v1/orgs)
	if len(parts) > 0 {
		entityType = parts[len(parts)-1]
	}
	return
}

func Audit(db *sqlx.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			method := c.Request().Method
			if method != "POST" && method != "PUT" && method != "DELETE" {
				return err
			}

			status := c.Response().Status
			if status < 200 || status >= 300 {
				return err
			}

			var userID *string
			if uid, ok := c.Get("user_id").(uuid.UUID); ok {
				s := uid.String()
				userID = &s
			}

			action := method + " " + c.Path()

			entityType, entityID := extractAuditEntity(c.Request().URL.Path)

			// Allow handlers to override entity_id (e.g. for create routes
			// where the new resource ID isn't in the URL path).
			if eid, ok := c.Get("audit_entity_id").(string); ok && eid != "" {
				entityID = eid
			}

			var oldValue *json.RawMessage
			if ov, ok := c.Get("audit_old_value").(json.RawMessage); ok {
				oldValue = &ov
			}

			var newValue *json.RawMessage
			if nv, ok := c.Get("audit_new_value").(json.RawMessage); ok {
				newValue = &nv
			}

			reqID, _ := c.Get("request_id").(string)
			ip := c.RealIP()

			auditLog := &model.AuditLog{
				UserID:     userID,
				Action:     action,
				EntityType: entityType,
				EntityID:   entityID,
				OldValue:   oldValue,
				NewValue:   newValue,
				IPAddress:  &ip,
				RequestID:  &reqID,
				CreatedAt:  time.Now(),
			}

			ctx := c.Request().Context()
			if auditErr := repo.CreateAuditLog(ctx, db, auditLog); auditErr != nil {
				log.Error().Err(auditErr).Msg("failed to create audit log")
			}

			return err
		}
	}
}
