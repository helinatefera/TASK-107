package handler

import (
	"net/http"
	"strconv"

	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// enforceVersionOrgOwnership resolves a version to its template and checks org ownership.
func enforceVersionOrgOwnership(c echo.Context, db *sqlx.DB, versionID uuid.UUID) error {
	version, err := service.GetVersion(c.Request().Context(), db, versionID)
	if err != nil {
		return err
	}
	tmpl, err := service.GetPriceTemplate(c.Request().Context(), db, version.TemplateID)
	if err != nil {
		return err
	}
	return enforceOrgOwnership(c, db, tmpl.OrgID)
}

func CreateTemplateHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreatePriceTemplateRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		// Derive org from the referenced station or device.
		ctx := c.Request().Context()
		var ownerOrgID uuid.UUID
		switch {
		case req.StationID != nil:
			station, err := service.GetStation(ctx, db, *req.StationID)
			if err != nil {
				return err
			}
			ownerOrgID = station.OrgID
		case req.DeviceID != nil:
			device, err := service.GetDevice(ctx, db, *req.DeviceID)
			if err != nil {
				return err
			}
			station, err := service.GetStation(ctx, db, device.StationID)
			if err != nil {
				return err
			}
			ownerOrgID = station.OrgID
		default:
			return echo.NewHTTPError(400, "station_id or device_id is required")
		}

		// Merchants can only create templates for their own org's stations/devices.
		if err := enforceOrgOwnership(c, db, ownerOrgID); err != nil {
			return err
		}

		tmpl, err := service.CreatePriceTemplate(c.Request().Context(), db, ownerOrgID, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, tmpl.ID.String())
		setAuditNewValue(c, tmpl)
		return c.JSON(http.StatusCreated, tmpl)
	}
}

func ListTemplatesHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))
		if limit <= 0 {
			limit = 50
		}
		if offset < 0 {
			offset = 0
		}

		orgID, err := getListOrgScope(c)
		if err != nil {
			return err
		}

		templates, err := service.ListPriceTemplates(c.Request().Context(), db, orgID, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, templates)
	}
}

func GetTemplateHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
		}

		tmpl, err := service.GetPriceTemplate(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		if err := enforceOrgOwnership(c, db, tmpl.OrgID); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, tmpl)
	}
}

func CreateVersionHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		templateID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
		}

		tmpl, err := service.GetPriceTemplate(c.Request().Context(), db, templateID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db, tmpl.OrgID); err != nil {
			return err
		}

		var req dto.CreateVersionRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}
		if err := req.ValidateRatesNonNegative(); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		version, err := service.CreateVersion(c.Request().Context(), db, templateID, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, version.ID.String())
		setAuditNewValue(c, version)
		return c.JSON(http.StatusCreated, version)
	}
}

func ListVersionsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		templateID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
		}

		tmpl, err := service.GetPriceTemplate(c.Request().Context(), db, templateID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db, tmpl.OrgID); err != nil {
			return err
		}

		versions, err := service.ListVersions(c.Request().Context(), db, templateID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, versions)
	}
}

func GetVersionHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
		}

		if err := enforceVersionOrgOwnership(c, db, id); err != nil {
			return err
		}

		version, err := service.GetVersion(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, version)
	}
}

func ActivateVersionHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
		}

		if err := enforceVersionOrgOwnership(c, db, id); err != nil {
			return err
		}

		var req dto.ActivateVersionRequest
		// Body is optional — if absent, effective_at defaults to now in the service
		c.Bind(&req)

		version, err := service.ActivateVersion(c.Request().Context(), db, id, req.EffectiveAt)
		if err != nil {
			return err
		}

		setAuditNewValue(c, version)
		return c.JSON(http.StatusOK, version)
	}
}

func DeactivateVersionHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
		}

		if err := enforceVersionOrgOwnership(c, db, id); err != nil {
			return err
		}

		if err := service.DeactivateVersion(c.Request().Context(), db, id); err != nil {
			return err
		}

		version, err := service.GetVersion(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		setAuditNewValue(c, version)
		return c.JSON(http.StatusOK, map[string]string{"status": "deactivated"})
	}
}

func RollbackVersionHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
		}

		if err := enforceVersionOrgOwnership(c, db, id); err != nil {
			return err
		}

		version, err := service.RollbackVersion(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		setAuditNewValue(c, version)
		return c.JSON(http.StatusCreated, version)
	}
}

func CreateTOURuleHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		versionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
		}

		if err := enforceVersionOrgOwnership(c, db, versionID); err != nil {
			return err
		}

		var req dto.CreateTOURuleRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}
		if err := req.ValidateRatesNonNegative(); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		rule, err := service.CreateTOURule(c.Request().Context(), db, versionID, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, rule)
		return c.JSON(http.StatusCreated, rule)
	}
}

func ListTOURulesHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		versionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
		}

		if err := enforceVersionOrgOwnership(c, db, versionID); err != nil {
			return err
		}

		rules, err := service.ListTOURules(c.Request().Context(), db, versionID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rules)
	}
}

func DeleteTOURuleHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid rule id")
		}

		rule, err := service.GetTOURule(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceVersionOrgOwnership(c, db, rule.VersionID); err != nil {
			return err
		}
		setAuditOldValue(c, rule)

		if err := service.DeleteTOURule(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}
