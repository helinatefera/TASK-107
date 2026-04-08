package handler

import (
	"net/http"
	"strconv"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// enforceSelfOrAdmin returns an error if the caller is not an administrator
// and is not accessing their own record.
func enforceSelfOrAdmin(c echo.Context, targetID uuid.UUID) error {
	role := getRole(c)
	if role == "administrator" {
		return nil
	}
	callerID := getUserID(c)
	if callerID != targetID {
		return apperror.ErrForbidden
	}
	return nil
}

func ListUsersHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))
		if limit <= 0 {
			limit = 50
		}
		if offset < 0 {
			offset = 0
		}

		users, err := service.ListUsers(c.Request().Context(), db, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, users)
	}
}

func GetCurrentUserHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := getUserID(c)

		user, err := service.GetUser(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, user)
	}
}

func GetUserHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}

		if err := enforceSelfOrAdmin(c, id); err != nil {
			return err
		}

		user, err := service.GetUser(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, user)
	}
}

func UpdateUserHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}

		if err := enforceSelfOrAdmin(c, id); err != nil {
			return err
		}

		var req dto.UpdateUserRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		user, err := service.UpdateUser(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, user)
	}
}

func UpdateUserOrgHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}

		old, err := service.GetUser(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		setAuditOldValue(c, map[string]interface{}{"id": id, "org_id": old.OrgID})

		var req struct {
			OrgID *uuid.UUID `json:"org_id"`
		}
		if err := c.Bind(&req); err != nil {
			return err
		}

		if err := service.SetUserOrg(c.Request().Context(), db, id, req.OrgID); err != nil {
			return err
		}

		setAuditNewValue(c, map[string]interface{}{"id": id, "org_id": req.OrgID})
		return c.JSON(http.StatusOK, map[string]string{"status": "org_updated"})
	}
}

func UpdateRoleHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}

		old, err := service.GetUser(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		setAuditOldValue(c, map[string]interface{}{"id": id, "role": old.Role})

		var req dto.UpdateRoleRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := service.UpdateUserRole(c.Request().Context(), db, id, req.Role); err != nil {
			return err
		}

		setAuditNewValue(c, map[string]interface{}{"id": id, "role": req.Role})
		return c.JSON(http.StatusOK, map[string]string{"status": "role_updated"})
	}
}

func DeleteUserHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}

		existing, err := service.GetUser(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		if err := service.DeleteUser(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func GetPermissionsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}

		if err := enforceSelfOrAdmin(c, id); err != nil {
			return err
		}

		user, err := service.GetUser(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		perms, err := service.GetEffectivePermissions(c.Request().Context(), db, id, user.Role)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, perms)
	}
}

func UpdatePermissionsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}

		user, err := service.GetUser(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		oldPerms, _ := service.GetEffectivePermissions(c.Request().Context(), db, id, user.Role)
		setAuditOldValue(c, map[string]interface{}{"id": id, "permissions": oldPerms})

		var req dto.UpdatePermissionsRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := service.UpdateUserPermissions(c.Request().Context(), db, id, req.Permissions); err != nil {
			return err
		}

		setAuditNewValue(c, map[string]interface{}{"id": id, "permissions": req.Permissions})
		return c.JSON(http.StatusOK, map[string]string{"status": "permissions_updated"})
	}
}
