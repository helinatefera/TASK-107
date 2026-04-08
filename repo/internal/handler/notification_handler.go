package handler

import (
	"net/http"
	"strconv"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/repo"
	"github.com/chargeops/api/internal/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func CreateNotificationTemplateHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateNotificationTemplateRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		tmpl, err := service.CreateNotificationTemplate(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, tmpl.ID.String())
		setAuditNewValue(c, tmpl)
		return c.JSON(http.StatusCreated, tmpl)
	}
}

func ListNotificationTemplatesHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		templates, err := service.ListNotificationTemplates(c.Request().Context(), db)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, templates)
	}
}

func ListSubscriptionsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := getUserID(c)

		subs, err := service.GetSubscriptions(c.Request().Context(), db, userID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, subs)
	}
}

func UpdateSubscriptionHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		templateID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid template id")
		}

		var req dto.UpdateSubscriptionRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		userID := getUserID(c)

		if err := service.UpdateSubscription(c.Request().Context(), db, userID, templateID, req.OptedOut); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
	}
}

func SendNotificationHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.SendNotificationRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := service.SendNotification(c.Request().Context(), db, &req); err != nil {
			return err
		}

		setAuditNewValue(c, req)
		return c.JSON(http.StatusCreated, map[string]string{"status": "queued"})
	}
}

func ListInboxHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))
		if limit <= 0 {
			limit = 50
		}
		if offset < 0 {
			offset = 0
		}

		userID := getUserID(c)

		messages, err := service.GetInbox(c.Request().Context(), db, userID, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, messages)
	}
}

func MarkReadHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid message id")
		}

		msg, err := repo.GetMessage(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		callerID := getUserID(c)
		if msg.UserID != callerID {
			return apperror.ErrForbidden
		}

		if err := service.MarkRead(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "read"})
	}
}

func MarkDismissHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid message id")
		}

		msg, err := repo.GetMessage(c.Request().Context(), db, id)
		if err != nil {
			return err
		}

		callerID := getUserID(c)
		if msg.UserID != callerID {
			return apperror.ErrForbidden
		}

		if err := service.MarkDismissed(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "dismissed"})
	}
}

func GetDeliveryStatsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var templateID *uuid.UUID
		if raw := c.QueryParam("template_id"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid template_id")
			}
			templateID = &id
		}

		stats, err := service.GetDeliveryStats(c.Request().Context(), db, templateID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, stats)
	}
}
