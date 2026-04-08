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

func CreateOrderHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateOrderRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		// Input validation: time range and energy
		if !req.StartTime.Before(req.EndTime) {
			return echo.NewHTTPError(400, "end_time must be after start_time")
		}
		if req.EnergyKWh.IsNegative() || req.EnergyKWh.IsZero() {
			return echo.NewHTTPError(400, "energy_kwh must be positive")
		}

		userID := getUserID(c)
		role := getRole(c)
		callerOrgID := getOrgID(c)
		isAdmin := role == "administrator"

		snapshot, err := service.CreateOrderSnapshot(c.Request().Context(), db, userID, &req, callerOrgID, isAdmin)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, snapshot)
	}
}

func ListOrdersHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
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
		var userID *uuid.UUID
		if role == "merchant" || role == "user" {
			id := getUserID(c)
			userID = &id
		}

		orders, err := service.ListOrderSnapshots(c.Request().Context(), db, userID, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orders)
	}
}

func GetOrderHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid order id")
		}

		order, err := service.GetOrderSnapshot(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		role := getRole(c)
		if role != "administrator" {
			callerID := getUserID(c)
			if order.UserID != callerID {
				return apperror.ErrForbidden
			}
		}

		return c.JSON(http.StatusOK, order)
	}
}

func RecalculateOrderHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid order id")
		}

		result, err := service.RecalculateOrder(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, result)
	}
}
