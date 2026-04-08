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

func CreateSupplierHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateSupplierRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := enforceOrOverrideOrgID(c, &req.OrgID); err != nil {
			return err
		}

		s, err := service.CreateSupplier(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		role := getRole(c)
		service.MaskSupplier(s, role)
		setAuditEntityID(c, s.ID.String())
		setAuditNewValue(c, s)
		return c.JSON(http.StatusCreated, s)
	}
}

func GetSupplierHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid supplier id")
		}

		s, err := service.GetSupplier(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		if s.OrgID != nil {
			if err := enforceOrgOwnership(c, db,*s.OrgID); err != nil {
				return err
			}
		}

		role := getRole(c)
		service.MaskSupplier(s, role)
		return c.JSON(http.StatusOK, s)
	}
}

func ListSuppliersHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
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
		suppliers, err := service.ListSuppliers(c.Request().Context(), db, orgID, limit, offset)
		if err != nil {
			return err
		}

		role := getRole(c)
		for i := range suppliers {
			service.MaskSupplier(&suppliers[i], role)
		}

		return c.JSON(http.StatusOK, suppliers)
	}
}

func UpdateSupplierHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid supplier id")
		}

		old, err := service.GetSupplier(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if old.OrgID != nil {
			if err := enforceOrgOwnership(c, db,*old.OrgID); err != nil {
				return err
			}
		}
		role := getRole(c)
		maskedOld := *old
		service.MaskSupplier(&maskedOld, role)
		setAuditOldValue(c, &maskedOld)

		var req dto.UpdateSupplierRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		s, err := service.UpdateSupplier(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		service.MaskSupplier(s, role)
		setAuditNewValue(c, s)
		return c.JSON(http.StatusOK, s)
	}
}

func CreateCarrierHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateCarrierRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := enforceOrOverrideOrgID(c, &req.OrgID); err != nil {
			return err
		}

		carrier, err := service.CreateCarrier(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		role := getRole(c)
		service.MaskCarrier(carrier, role)
		setAuditEntityID(c, carrier.ID.String())
		setAuditNewValue(c, carrier)
		return c.JSON(http.StatusCreated, carrier)
	}
}

func GetCarrierHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid carrier id")
		}

		carrier, err := service.GetCarrier(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		if carrier.OrgID != nil {
			if err := enforceOrgOwnership(c, db,*carrier.OrgID); err != nil {
				return err
			}
		}

		role := getRole(c)
		service.MaskCarrier(carrier, role)
		return c.JSON(http.StatusOK, carrier)
	}
}

func ListCarriersHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
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
		carriers, err := service.ListCarriers(c.Request().Context(), db, orgID, limit, offset)
		if err != nil {
			return err
		}

		role := getRole(c)
		for i := range carriers {
			service.MaskCarrier(&carriers[i], role)
		}

		return c.JSON(http.StatusOK, carriers)
	}
}

func UpdateCarrierHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid carrier id")
		}

		old, err := service.GetCarrier(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if old.OrgID != nil {
			if err := enforceOrgOwnership(c, db,*old.OrgID); err != nil {
				return err
			}
		}
		role := getRole(c)
		maskedOld := *old
		service.MaskCarrier(&maskedOld, role)
		setAuditOldValue(c, &maskedOld)

		var req dto.UpdateCarrierRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		carrier, err := service.UpdateCarrier(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		service.MaskCarrier(carrier, role)
		setAuditNewValue(c, carrier)
		return c.JSON(http.StatusOK, carrier)
	}
}
