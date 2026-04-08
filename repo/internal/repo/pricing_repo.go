package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateTemplate(ctx context.Context, db sqlx.ExtContext, t *model.PriceTemplate) error {
	query := `INSERT INTO price_templates (id, org_id, name, station_id, device_id, created_at)
		VALUES (:id, :org_id, :name, :station_id, :device_id, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, t)
	return err
}

func GetTemplate(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.PriceTemplate, error) {
	var t model.PriceTemplate
	err := sqlx.GetContext(ctx, db, &t, "SELECT * FROM price_templates WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &t, err
}

func ListTemplates(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.PriceTemplate, error) {
	var templates []model.PriceTemplate
	if orgID != nil {
		err := sqlx.SelectContext(ctx, db, &templates, `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id) SELECT pt.* FROM price_templates pt JOIN org_tree ot ON pt.org_id = ot.id ORDER BY pt.created_at DESC LIMIT $2 OFFSET $3`, *orgID, limit, offset)
		return templates, err
	}
	err := sqlx.SelectContext(ctx, db, &templates, "SELECT * FROM price_templates ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return templates, err
}

func CreateVersion(ctx context.Context, db sqlx.ExtContext, v *model.PriceTemplateVersion) error {
	query := `INSERT INTO price_template_versions (id, template_id, version_number, energy_rate, duration_rate, service_fee, apply_tax, tax_rate, status, effective_at, cloned_from_version_id, created_at)
		VALUES (:id, :template_id, :version_number, :energy_rate, :duration_rate, :service_fee, :apply_tax, :tax_rate, :status, :effective_at, :cloned_from_version_id, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, v)
	return err
}

func GetVersion(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.PriceTemplateVersion, error) {
	var v model.PriceTemplateVersion
	err := sqlx.GetContext(ctx, db, &v, "SELECT * FROM price_template_versions WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &v, err
}

func ListVersions(ctx context.Context, db sqlx.ExtContext, templateID uuid.UUID) ([]model.PriceTemplateVersion, error) {
	var versions []model.PriceTemplateVersion
	err := sqlx.SelectContext(ctx, db, &versions, "SELECT * FROM price_template_versions WHERE template_id = $1 ORDER BY version_number DESC", templateID)
	return versions, err
}

func GetMaxVersionNumber(ctx context.Context, db sqlx.ExtContext, templateID uuid.UUID) (int, error) {
	var max int
	err := sqlx.GetContext(ctx, db, &max, "SELECT COALESCE(MAX(version_number), 0) FROM price_template_versions WHERE template_id = $1", templateID)
	return max, err
}

func ActivateVersion(ctx context.Context, db sqlx.ExtContext, versionID uuid.UUID, effectiveAt time.Time) error {
	_, err := db.ExecContext(ctx, "UPDATE price_template_versions SET status = 'active', effective_at = $1 WHERE id = $2", effectiveAt, versionID)
	return err
}

func DeactivateCurrentActive(ctx context.Context, db sqlx.ExtContext, templateID uuid.UUID) error {
	_, err := db.ExecContext(ctx, "UPDATE price_template_versions SET status = 'inactive' WHERE template_id = $1 AND status = 'active'", templateID)
	return err
}

func DeactivateVersion(ctx context.Context, db sqlx.ExtContext, versionID uuid.UUID) error {
	_, err := db.ExecContext(ctx, "UPDATE price_template_versions SET status = 'inactive' WHERE id = $1", versionID)
	return err
}

func GetActiveVersionForDevice(ctx context.Context, db sqlx.ExtContext, deviceID uuid.UUID, at time.Time) (*model.PriceTemplateVersion, error) {
	var v model.PriceTemplateVersion
	query := `SELECT pv.* FROM price_template_versions pv
		JOIN price_templates pt ON pt.id = pv.template_id
		WHERE pt.device_id = $1 AND pv.status = 'active' AND pv.effective_at <= $2
		ORDER BY pv.effective_at DESC LIMIT 1`
	err := sqlx.GetContext(ctx, db, &v, query, deviceID, at)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &v, err
}

func GetActiveVersionForStation(ctx context.Context, db sqlx.ExtContext, stationID uuid.UUID, at time.Time) (*model.PriceTemplateVersion, error) {
	var v model.PriceTemplateVersion
	query := `SELECT pv.* FROM price_template_versions pv
		JOIN price_templates pt ON pt.id = pv.template_id
		WHERE pt.station_id = $1 AND pv.status = 'active' AND pv.effective_at <= $2
		ORDER BY pv.effective_at DESC LIMIT 1`
	err := sqlx.GetContext(ctx, db, &v, query, stationID, at)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &v, err
}

func GetTOURule(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.TOURule, error) {
	var r model.TOURule
	err := sqlx.GetContext(ctx, db, &r, "SELECT * FROM tou_rules WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &r, err
}

func CreateTOURule(ctx context.Context, db sqlx.ExtContext, r *model.TOURule) error {
	query := `INSERT INTO tou_rules (id, version_id, day_type, start_time, end_time, energy_rate, duration_rate)
		VALUES (:id, :version_id, :day_type, :start_time, :end_time, :energy_rate, :duration_rate)`
	_, err := sqlx.NamedExecContext(ctx, db, query, r)
	return err
}

func ListTOURules(ctx context.Context, db sqlx.ExtContext, versionID uuid.UUID) ([]model.TOURule, error) {
	var rules []model.TOURule
	err := sqlx.SelectContext(ctx, db, &rules, "SELECT * FROM tou_rules WHERE version_id = $1 ORDER BY day_type, start_time", versionID)
	return rules, err
}

func DeleteTOURule(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM tou_rules WHERE id = $1", id)
	return err
}

func CreateOrderSnapshot(ctx context.Context, db sqlx.ExtContext, o *model.OrderSnapshot) error {
	query := `INSERT INTO order_snapshots (id, order_id, user_id, device_id, station_id, version_id, energy_rate, duration_rate, service_fee, tax_rate, tou_applied, energy_kwh, duration_min, subtotal, tax_amount, total, order_start, order_end, created_at)
		VALUES (:id, :order_id, :user_id, :device_id, :station_id, :version_id, :energy_rate, :duration_rate, :service_fee, :tax_rate, :tou_applied, :energy_kwh, :duration_min, :subtotal, :tax_amount, :total, :order_start, :order_end, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, o)
	return err
}

func GetOrderSnapshot(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.OrderSnapshot, error) {
	var o model.OrderSnapshot
	err := sqlx.GetContext(ctx, db, &o, "SELECT id, order_id, user_id, device_id, station_id, version_id, energy_rate, duration_rate, service_fee, tax_rate, COALESCE(tou_applied, 'null'::jsonb) AS tou_applied, energy_kwh, duration_min, subtotal, tax_amount, total, order_start, order_end, created_at FROM order_snapshots WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &o, err
}

func ListOrderSnapshots(ctx context.Context, db sqlx.ExtContext, userID *uuid.UUID, limit, offset int) ([]model.OrderSnapshot, error) {
	var snapshots []model.OrderSnapshot
	orderCols := "id, order_id, user_id, device_id, station_id, version_id, energy_rate, duration_rate, service_fee, tax_rate, COALESCE(tou_applied, 'null'::jsonb) AS tou_applied, energy_kwh, duration_min, subtotal, tax_amount, total, order_start, order_end, created_at"
	if userID != nil {
		err := sqlx.SelectContext(ctx, db, &snapshots, "SELECT "+orderCols+" FROM order_snapshots WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", *userID, limit, offset)
		return snapshots, err
	}
	err := sqlx.SelectContext(ctx, db, &snapshots, "SELECT "+orderCols+" FROM order_snapshots ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return snapshots, err
}
