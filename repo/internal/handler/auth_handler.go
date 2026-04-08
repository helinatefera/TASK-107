package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/chargeops/api/internal/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func setAuditOldValue(c echo.Context, v interface{}) {
	if b, err := json.Marshal(v); err == nil {
		c.Set("audit_old_value", json.RawMessage(b))
	}
}

func setAuditNewValue(c echo.Context, v interface{}) {
	if b, err := json.Marshal(v); err == nil {
		c.Set("audit_new_value", json.RawMessage(b))
	}
}

func setAuditEntityID(c echo.Context, id string) {
	c.Set("audit_entity_id", id)
}

func getUserID(c echo.Context) uuid.UUID   { return c.Get("user_id").(uuid.UUID) }
func getRole(c echo.Context) string        { return c.Get("role").(string) }
func getOrgID(c echo.Context) *uuid.UUID {
	if v := c.Get("org_id"); v != nil {
		id := v.(*uuid.UUID)
		return id
	}
	return nil
}
func getSession(c echo.Context) *model.Session { return c.Get("session").(*model.Session) }

// getListOrgScope returns orgID for list operations: nil for admins (all),
// caller's org for merchants/users, or error if non-admin has no org.
func getListOrgScope(c echo.Context) (*uuid.UUID, error) {
	role := getRole(c)
	if role == "administrator" {
		return nil, nil // admins see everything
	}
	orgID := getOrgID(c)
	if orgID == nil {
		return nil, apperror.ErrForbidden
	}
	return orgID, nil
}

// enforceOrOverrideOrgID forces merchants to use their own org_id, admins pass through.
func enforceOrOverrideOrgID(c echo.Context, orgID *uuid.UUID) error {
	role := getRole(c)
	if role == "administrator" {
		return nil
	}
	callerOrgID := getOrgID(c)
	if callerOrgID == nil {
		return apperror.ErrForbidden
	}
	*orgID = *callerOrgID
	return nil
}

// enforceOrgOwnership checks that the resource's org is the caller's org or a
// descendant at any depth in the org hierarchy. Admins bypass this check.
func enforceOrgOwnership(c echo.Context, db sqlx.ExtContext, resourceOrgID uuid.UUID) error {
	role := getRole(c)
	if role == "administrator" {
		return nil
	}
	callerOrgID := getOrgID(c)
	if callerOrgID == nil {
		return apperror.ErrForbidden
	}
	if *callerOrgID == resourceOrgID {
		return nil
	}
	accessible, err := repo.IsOrgAccessible(c.Request().Context(), db, *callerOrgID, resourceOrgID)
	if err == nil && accessible {
		return nil
	}
	return apperror.ErrForbidden
}

func RegisterHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.RegisterRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		user, err := service.Register(c.Request().Context(), db, &req, cfg)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, user)
	}
}

func LoginHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.LoginRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		token, session, err := service.Login(c.Request().Context(), db, &req, cfg)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.LoginResponse{
			Token:     token,
			ExpiresAt: session.IdleExpiresAt.Format(time.RFC3339),
		})
	}
}

func LogoutHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := getSession(c)

		if err := service.Logout(c.Request().Context(), db, session.ID); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func RefreshHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.RefreshRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		session := getSession(c)

		if err := service.RefreshSession(c.Request().Context(), db, session, req.DeviceID, cfg); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "refreshed"})
	}
}

func RecoverHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.RecoverRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		token, err := service.RequestRecovery(c.Request().Context(), db, req.Email, cfg)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.RecoverResponse{Token: token})
	}
}

func ResetPasswordHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.ResetPasswordRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := service.ResetPassword(c.Request().Context(), db, &req, cfg); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "password_reset"})
	}
}
