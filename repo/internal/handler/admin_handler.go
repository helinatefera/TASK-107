package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// validateConfigValue ensures known config keys have valid values.
func validateConfigValue(key, value string) error {
	switch key {
	case "tax_rate":
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "tax_rate must be a valid decimal number")
		}
		if rate < 0 || rate > 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "tax_rate must be between 0 and 1")
		}
	}
	return nil
}

func ListAuditLogsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		entityType := c.QueryParam("entity_type")
		entityID := c.QueryParam("entity_id")

		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))
		if limit <= 0 {
			limit = 50
		}
		if offset < 0 {
			offset = 0
		}

		logs, err := repo.ListAuditLogs(c.Request().Context(), db, entityType, entityID, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, logs)
	}
}

func GetConfigHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		configs, err := repo.ListConfigs(c.Request().Context(), db)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, configs)
	}
}

func UpdateConfigHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.Param("key")

		old, _ := repo.GetConfig(c.Request().Context(), db, key)
		if old != nil {
			setAuditOldValue(c, old)
		}

		var body struct {
			Value string `json:"value" validate:"required"`
		}
		if err := c.Bind(&body); err != nil {
			return err
		}
		if err := c.Validate(body); err != nil {
			return err
		}

		// Validate known config keys
		if err := validateConfigValue(key, body.Value); err != nil {
			return err
		}

		appCfg := &model.AppConfig{
			Key:       key,
			Value:     body.Value,
			UpdatedAt: time.Now().UTC(),
		}

		if err := repo.UpsertConfig(c.Request().Context(), db, appCfg); err != nil {
			return err
		}

		setAuditNewValue(c, appCfg)
		return c.JSON(http.StatusOK, appCfg)
	}
}

func ListMetricsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))
		if limit <= 0 {
			limit = 50
		}
		if offset < 0 {
			offset = 0
		}

		metrics, err := repo.ListMetrics(c.Request().Context(), db, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, metrics)
	}
}

// ExportMetricsHandler exports request metrics in CSV or JSON format.
//
// Query parameters:
//   - format: "csv" (default) or "json"
//   - since: RFC3339 timestamp (default: 24 hours ago)
//   - until: RFC3339 timestamp (default: now)
//   - path: filter by request path (e.g., "/api/v1/auth/login")
//   - status_code: filter by HTTP status code (e.g., 200)
//
// Exported fields: id, method, path, status_code, latency_ms, recorded_at
func ExportMetricsHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse filters
		filter := repo.MetricsFilter{
			Since: time.Now().UTC().Add(-24 * time.Hour),
		}
		if raw := c.QueryParam("since"); raw != "" {
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid since parameter, use RFC3339 format")
			}
			filter.Since = parsed
		}
		if raw := c.QueryParam("until"); raw != "" {
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid until parameter, use RFC3339 format")
			}
			filter.Until = parsed
		}
		if p := c.QueryParam("path"); p != "" {
			filter.Path = p
		}
		if sc := c.QueryParam("status_code"); sc != "" {
			code, err := strconv.Atoi(sc)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid status_code parameter")
			}
			filter.StatusCode = code
		}

		metrics, err := repo.ExportMetrics(c.Request().Context(), db, filter)
		if err != nil {
			return err
		}

		format := c.QueryParam("format")
		if format == "json" {
			return exportMetricsJSON(c, metrics)
		}
		return exportMetricsCSV(c, metrics)
	}
}

func exportMetricsCSV(c echo.Context, metrics []model.RequestMetric) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=metrics.csv")
	c.Response().WriteHeader(http.StatusOK)

	writer := csv.NewWriter(c.Response())
	defer writer.Flush()

	if err := writer.Write([]string{"id", "method", "path", "status_code", "latency_ms", "recorded_at"}); err != nil {
		return err
	}

	for _, m := range metrics {
		row := []string{
			fmt.Sprintf("%d", m.ID),
			m.Method,
			m.Path,
			strconv.Itoa(m.StatusCode),
			strconv.Itoa(m.LatencyMs),
			m.RecordedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func exportMetricsJSON(c echo.Context, metrics []model.RequestMetric) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=metrics.json")
	c.Response().WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.Response())
	encoder.SetIndent("", "  ")
	return encoder.Encode(metrics)
}

func HealthHandler(db *sqlx.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}
