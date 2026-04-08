package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/chargeops/api/internal/model"
	"github.com/jmoiron/sqlx"
)

func CreateAuditLog(ctx context.Context, db sqlx.ExtContext, log *model.AuditLog) error {
	query := `INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_value, new_value, ip_address, request_id, created_at)
		VALUES (:user_id, :action, :entity_type, :entity_id, :old_value, :new_value, :ip_address, :request_id, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, log)
	return err
}

func ListAuditLogs(ctx context.Context, db sqlx.ExtContext, entityType string, entityID string, limit, offset int) ([]model.AuditLog, error) {
	var logs []model.AuditLog

	if entityType != "" && entityID != "" {
		err := sqlx.SelectContext(ctx, db, &logs,
			"SELECT * FROM audit_logs WHERE entity_type = $1 AND entity_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4",
			entityType, entityID, limit, offset)
		return logs, err
	}
	if entityType != "" {
		err := sqlx.SelectContext(ctx, db, &logs,
			"SELECT * FROM audit_logs WHERE entity_type = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			entityType, limit, offset)
		return logs, err
	}

	err := sqlx.SelectContext(ctx, db, &logs,
		"SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset)
	return logs, err
}

func GetConfig(ctx context.Context, db sqlx.ExtContext, key string) (*model.AppConfig, error) {
	var cfg model.AppConfig
	err := sqlx.GetContext(ctx, db, &cfg, "SELECT * FROM app_config WHERE key = $1", key)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ListConfigs(ctx context.Context, db sqlx.ExtContext) ([]model.AppConfig, error) {
	var configs []model.AppConfig
	err := sqlx.SelectContext(ctx, db, &configs, "SELECT * FROM app_config ORDER BY key")
	return configs, err
}

func UpsertConfig(ctx context.Context, db sqlx.ExtContext, cfg *model.AppConfig) error {
	query := `INSERT INTO app_config (key, value, updated_at)
		VALUES (:key, :value, :updated_at)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`
	_, err := sqlx.NamedExecContext(ctx, db, query, cfg)
	return err
}

func InsertMetric(ctx context.Context, db sqlx.ExtContext, m *model.RequestMetric) error {
	query := `INSERT INTO request_metrics (method, path, status_code, latency_ms, recorded_at)
		VALUES (:method, :path, :status_code, :latency_ms, :recorded_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, m)
	return err
}

func ListMetrics(ctx context.Context, db sqlx.ExtContext, limit, offset int) ([]model.RequestMetric, error) {
	var metrics []model.RequestMetric
	err := sqlx.SelectContext(ctx, db, &metrics,
		"SELECT * FROM request_metrics ORDER BY recorded_at DESC LIMIT $1 OFFSET $2",
		limit, offset)
	return metrics, err
}

type MetricsFilter struct {
	Since      time.Time
	Until      time.Time
	Path       string
	StatusCode int
}

func ExportMetrics(ctx context.Context, db sqlx.ExtContext, f MetricsFilter) ([]model.RequestMetric, error) {
	var metrics []model.RequestMetric
	query := "SELECT * FROM request_metrics WHERE recorded_at >= $1"
	args := []interface{}{f.Since}
	argN := 2

	if !f.Until.IsZero() {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argN)
		args = append(args, f.Until)
		argN++
	}
	if f.Path != "" {
		query += fmt.Sprintf(" AND path = $%d", argN)
		args = append(args, f.Path)
		argN++
	}
	if f.StatusCode > 0 {
		query += fmt.Sprintf(" AND status_code = $%d", argN)
		args = append(args, f.StatusCode)
		argN++
	}

	query += " ORDER BY recorded_at"
	err := sqlx.SelectContext(ctx, db, &metrics, query, args...)
	return metrics, err
}

func PurgeMetricsOlderThan(ctx context.Context, db sqlx.ExtContext, before time.Time) (int64, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM request_metrics WHERE recorded_at < $1", before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
