package handler

import (
	"net/http"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// --- Carousel Handlers ---

func CreateCarouselHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateCarouselRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := enforceOrOverrideOrgID(c, &req.OrgID); err != nil {
			return err
		}

		carousel, err := service.CreateCarousel(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, carousel.ID.String())
		setAuditNewValue(c, carousel)
		return c.JSON(http.StatusCreated, carousel)
	}
}

func ListCarouselsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := getRole(c)
		orgID := getOrgID(c)
		if role == "merchant" && orgID == nil {
			return apperror.ErrForbidden
		}

		carousels, err := service.ListCarousels(c.Request().Context(), db, role, orgID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, carousels)
	}
}

func UpdateCarouselHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid carousel id")
		}

		old, err := service.GetCarousel(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,old.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, old)

		var req dto.UpdateCarouselRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		carousel, err := service.UpdateCarousel(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, carousel)
		return c.JSON(http.StatusOK, carousel)
	}
}

func DeleteCarouselHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid carousel id")
		}

		existing, err := service.GetCarousel(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,existing.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		if err := service.DeleteCarousel(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// --- Campaign Handlers ---

func CreateCampaignHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateCampaignRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := enforceOrOverrideOrgID(c, &req.OrgID); err != nil {
			return err
		}

		campaign, err := service.CreateCampaign(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, campaign.ID.String())
		setAuditNewValue(c, campaign)
		return c.JSON(http.StatusCreated, campaign)
	}
}

func ListCampaignsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := getRole(c)

		orgID := getOrgID(c)
		if role == "merchant" && orgID == nil {
			return apperror.ErrForbidden
		}
		campaigns, err := service.ListCampaigns(c.Request().Context(), db, role, orgID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, campaigns)
	}
}

func UpdateCampaignHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid campaign id")
		}

		old, err := service.GetCampaign(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,old.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, old)

		var req dto.UpdateCampaignRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		campaign, err := service.UpdateCampaign(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, campaign)
		return c.JSON(http.StatusOK, campaign)
	}
}

func DeleteCampaignHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid campaign id")
		}

		existing, err := service.GetCampaign(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,existing.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		if err := service.DeleteCampaign(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// --- Ranking Handlers ---

func CreateRankingHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateRankingRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		if err := enforceOrOverrideOrgID(c, &req.OrgID); err != nil {
			return err
		}

		ranking, err := service.CreateRanking(c.Request().Context(), db, &req)
		if err != nil {
			return err
		}

		setAuditEntityID(c, ranking.ID.String())
		setAuditNewValue(c, ranking)
		return c.JSON(http.StatusCreated, ranking)
	}
}

func ListRankingsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := getRole(c)

		orgID := getOrgID(c)
		if role == "merchant" && orgID == nil {
			return apperror.ErrForbidden
		}
		rankings, err := service.ListRankings(c.Request().Context(), db, role, orgID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rankings)
	}
}

func UpdateRankingHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ranking id")
		}

		old, err := service.GetRanking(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,old.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, old)

		var req dto.UpdateRankingRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		ranking, err := service.UpdateRanking(c.Request().Context(), db, id, &req)
		if err != nil {
			return err
		}

		setAuditNewValue(c, ranking)
		return c.JSON(http.StatusOK, ranking)
	}
}

func DeleteRankingHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ranking id")
		}

		existing, err := service.GetRanking(c.Request().Context(), db, id)
		if err != nil {
			return err
		}
		if err := enforceOrgOwnership(c, db,existing.OrgID); err != nil {
			return err
		}
		setAuditOldValue(c, existing)

		if err := service.DeleteRanking(c.Request().Context(), db, id); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}
