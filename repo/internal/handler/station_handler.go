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

func CreateStationHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateStationRequest
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

		station, err := service.CreateStation(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, station.ID.String())
		setAuditNewValue(c, station)
		return c.JSON(http.StatusCreated, station)
	}
}

func ListStationsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
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

		stations, err := service.ListStations(c.Request().Context(), db, orgID, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, stations)
	}
}

func GetStationHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid station id")
		}

		station, err := service.GetStation(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		if err := enforceOrgOwnership(c, db,station.OrgID); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, station)
	}
}

func UpdateStationHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid station id")
		}

		existing, err := service.GetStation(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,existing.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		var req dto.UpdateStationRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		station, err := service.UpdateStation(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, station)
		return c.JSON(http.StatusOK, station)
	}
}

func DeleteStationHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid station id")
		}

		existing, err := service.GetStation(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,existing.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		if err := service.DeleteStation(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func CreateDeviceHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		stationID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid station id")
		}

		station, err := service.GetStation(c.Request().Context(), db, stationID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,station.OrgID); err != nil {
			return err
		}

		var req dto.CreateDeviceRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		device, err := service.CreateDevice(c.Request().Context(), db, stationID, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, device.ID.String())
		setAuditNewValue(c, device)
		return c.JSON(http.StatusCreated, device)
	}
}

func ListDevicesHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		stationID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid station id")
		}

		station, err := service.GetStation(c.Request().Context(), db, stationID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,station.OrgID); err != nil {
			return err
		}

		devices, err := service.ListDevices(c.Request().Context(), db, stationID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, devices)
	}
}

func UpdateDeviceHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		deviceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid device id")
		}

		device, err := service.GetDevice(c.Request().Context(), db, deviceID)
		if err != nil {
			return err
		}
		station, err := service.GetStation(c.Request().Context(), db, device.StationID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,station.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, device)

		var req dto.UpdateDeviceRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		updated, err := service.UpdateDevice(c.Request().Context(), db, deviceID, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, updated)
		return c.JSON(http.StatusOK, updated)
	}
}

func DeleteDeviceHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		deviceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid device id")
		}

		device, err := service.GetDevice(c.Request().Context(), db, deviceID)
		if err != nil {
			return err
		}
		station, err := service.GetStation(c.Request().Context(), db, device.StationID)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,station.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, device)

		if err := service.DeleteDevice(c.Request().Context(), db, deviceID); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}
