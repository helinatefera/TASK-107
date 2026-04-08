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

func CreateOrgHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateOrgRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		org, err := service.CreateOrg(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		role := getRole(c)
		setAuditEntityID(c, org.ID.String())
		setAuditNewValue(c, service.MaskOrg(org, role))
		return c.JSON(http.StatusCreated, service.MaskOrg(org, role))
	}
}

func ListOrgsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))
		if limit <= 0 {
			limit = 50
		}
		if offset < 0 {
			offset = 0
		}

		role := getRole(c)
		orgID, err := getListOrgScope(c)
		if err != nil {
			return err
		}

		orgs, err := service.ListOrgs(c.Request().Context(), db, orgID, limit, offset)
		if err != nil {
			return err
		}

		result := make([]dto.OrgResponse, len(orgs))
		for i := range orgs {
			result[i] = service.MaskOrg(&orgs[i], role)
		}

		return c.JSON(http.StatusOK, result)
	}
}

func GetOrgHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid org id")
		}

		org, err := service.GetOrg(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		if err := enforceOrgOwnership(c, db,org.ID); err != nil {
			return err
		}

		role := getRole(c)
		return c.JSON(http.StatusOK, service.MaskOrg(org, role))
	}
}

func UpdateOrgHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid org id")
		}

		var req dto.UpdateOrgRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		old, err := service.GetOrg(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		role := getRole(c)
		setAuditOldValue(c, service.MaskOrg(old, role))

		org, err := service.UpdateOrg(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		masked := service.MaskOrg(org, role)
		setAuditNewValue(c, masked)
		return c.JSON(http.StatusOK, masked)
	}
}

func DeleteOrgHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid org id")
		}

		existing, err := service.GetOrg(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		role := getRole(c)
		setAuditOldValue(c, service.MaskOrg(existing, role))

		if err := service.DeleteOrg(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}
