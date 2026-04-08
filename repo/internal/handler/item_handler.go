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

func CreateCategoryHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateCategoryRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		cat, err := service.CreateCategory(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, cat.ID.String())
		setAuditNewValue(c, cat)
		return c.JSON(http.StatusCreated, cat)
	}
}

func ListCategoriesHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		cats, err := service.ListCategories(c.Request().Context(), db)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, cats)
	}
}

func CreateItemHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateItemRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		item, err := service.CreateItem(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, item.ID.String())
		setAuditNewValue(c, item)
		return c.JSON(http.StatusCreated, item)
	}
}

func GetItemHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid item id")
		}

		item, err := service.GetItem(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, item)
	}
}

func ListItemsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))
		if limit <= 0 {
			limit = 50
		}
		if offset < 0 {
			offset = 0
		}

		items, err := service.ListItems(c.Request().Context(), db, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, items)
	}
}

func UpdateItemHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid item id")
		}

		old, err := service.GetItem(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		setAuditOldValue(c, old)

		var req dto.UpdateItemRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		item, err := service.UpdateItem(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, item)
		return c.JSON(http.StatusOK, item)
	}
}

func DeleteItemHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid item id")
		}

		existing, err := service.GetItem(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		if err := service.DeleteItem(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func CreateUnitHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateUnitRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		unit, err := service.CreateUnit(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, unit.ID.String())
		setAuditNewValue(c, unit)
		return c.JSON(http.StatusCreated, unit)
	}
}

func ListUnitsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		units, err := service.ListUnits(c.Request().Context(), db)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, units)
	}
}

func CreateConversionHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateConversionRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		conv, err := service.CreateConversion(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, conv.ID.String())
		setAuditNewValue(c, conv)
		return c.JSON(http.StatusCreated, conv)
	}
}

func ListConversionsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		convs, err := service.ListConversions(c.Request().Context(), db)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, convs)
	}
}
