package handler

import (
	"net/http"
	"strconv"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/masking"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func maskWarehouseAddress(w *model.Warehouse, role string) {
	if w.Address != nil {
		masked := masking.MaskAddress(*w.Address, role)
		w.Address = &masked
	}
}

func CreateWarehouseHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateWarehouseRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		role := getRole(c)
		if role != "administrator" {
			callerOrgID := getOrgID(c)
			if callerOrgID == nil {
				return apperror.ErrForbidden
			}
			req.OrgID = *callerOrgID
		}

		w, err := service.CreateWarehouse(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		maskWarehouseAddress(w, role)
		setAuditEntityID(c, w.ID.String())
		setAuditNewValue(c, w)
		return c.JSON(http.StatusCreated, w)
	}
}

func ListWarehousesHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
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

		warehouses, err := service.ListWarehouses(c.Request().Context(), db, orgID, limit, offset)
		if err != nil {
			return err
		}

		role := getRole(c)
		for i := range warehouses {
			maskWarehouseAddress(&warehouses[i], role)
		}

		return c.JSON(http.StatusOK, warehouses)
	}
}

func GetWarehouseHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid warehouse id")
		}

		w, err := service.GetWarehouse(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		if err := enforceOrgOwnership(c, db,w.OrgID); err != nil {
			return err
		}

		role := getRole(c)
		maskWarehouseAddress(w, role)
		return c.JSON(http.StatusOK, w)
	}
}

func UpdateWarehouseHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid warehouse id")
		}

		existing, err := service.GetWarehouse(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,existing.OrgID); err != nil {
			return err
		}
		role := getRole(c)
		maskedOld := *existing
		maskWarehouseAddress(&maskedOld, role)
		setAuditOldValue(c, &maskedOld)

		var req dto.UpdateWarehouseRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		w, err := service.UpdateWarehouse(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		maskWarehouseAddress(w, role)
		setAuditNewValue(c, w)
		return c.JSON(http.StatusOK, w)
	}
}

func DeleteWarehouseHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid warehouse id")
		}

		existing, err := service.GetWarehouse(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,existing.OrgID); err != nil {
			return err
		}
		maskedOld := *existing
		maskWarehouseAddress(&maskedOld, getRole(c))
		setAuditOldValue(c, &maskedOld)

		if err := service.DeleteWarehouse(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func CreateZoneHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		warehouseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid warehouse id")
		}

		w, err := service.GetWarehouse(c.Request().Context(), db, warehouseID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,w.OrgID); err != nil {
			return err
		}

		var req dto.CreateZoneRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		zone, err := service.CreateZone(c.Request().Context(), db, warehouseID, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, zone.ID.String())
		setAuditNewValue(c, zone)
		return c.JSON(http.StatusCreated, zone)
	}
}

func ListZonesHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		warehouseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid warehouse id")
		}

		w, err := service.GetWarehouse(c.Request().Context(), db, warehouseID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,w.OrgID); err != nil {
			return err
		}

		zones, err := service.ListZones(c.Request().Context(), db, warehouseID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, zones)
	}
}

func CreateBinHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		zoneID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid zone id")
		}

		// Look up zone to get warehouse_id
		zone, err := service.GetZone(c.Request().Context(), db, zoneID)
		if err != nil {
			return err
		}
		warehouseID := zone.WarehouseID

		wh, err := service.GetWarehouse(c.Request().Context(), db, warehouseID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,wh.OrgID); err != nil {
			return err
		}

		var req dto.CreateBinRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		bin, err := service.CreateBin(c.Request().Context(), db, zoneID, warehouseID, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, bin.ID.String())
		setAuditNewValue(c, bin)
		return c.JSON(http.StatusCreated, bin)
	}
}

func ListBinsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		zoneID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid zone id")
		}

		zone, err := service.GetZone(c.Request().Context(), db, zoneID)
		if err != nil {
			return err
		}
		wh, err := service.GetWarehouse(c.Request().Context(), db, zone.WarehouseID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,wh.OrgID); err != nil {
			return err
		}

		bins, err := service.ListBins(c.Request().Context(), db, zoneID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, bins)
	}
}

func UpdateBinHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		binID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid bin id")
		}

		existing, err := service.GetBin(c.Request().Context(), db, binID)
		if err != nil {
			return err
		}
		wh, err := service.GetWarehouse(c.Request().Context(), db, existing.WarehouseID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,wh.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		var req dto.UpdateBinRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		bin, err := service.UpdateBin(c.Request().Context(), db, binID, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, bin)
		return c.JSON(http.StatusOK, bin)
	}
}

func DeleteBinHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		binID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid bin id")
		}

		existing, err := service.GetBin(c.Request().Context(), db, binID)
		if err != nil {
			return err
		}
		wh, err := service.GetWarehouse(c.Request().Context(), db, existing.WarehouseID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,wh.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		if err := service.DeleteBin(c.Request().Context(), db, binID); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}
