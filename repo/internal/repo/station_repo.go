package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateStation(ctx context.Context, db sqlx.ExtContext, s *model.Station) error {
	query := `INSERT INTO stations (id, org_id, name, location, timezone, created_at, updated_at)
		VALUES (:id, :org_id, :name, :location, :timezone, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, s)
	return err
}

func GetStation(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Station, error) {
	var s model.Station
	err := sqlx.GetContext(ctx, db, &s, "SELECT * FROM stations WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &s, err
}

func ListStations(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Station, error) {
	var stations []model.Station
	if orgID != nil {
		err := sqlx.SelectContext(ctx, db, &stations, `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id) SELECT s.* FROM stations s JOIN org_tree ot ON s.org_id = ot.id ORDER BY s.created_at DESC LIMIT $2 OFFSET $3`, *orgID, limit, offset)
		return stations, err
	}
	err := sqlx.SelectContext(ctx, db, &stations, "SELECT * FROM stations ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return stations, err
}

func UpdateStation(ctx context.Context, db sqlx.ExtContext, s *model.Station) error {
	query := `UPDATE stations SET org_id = $1, name = $2, location = $3, timezone = $4, updated_at = $5 WHERE id = $6`
	_, err := db.ExecContext(ctx, query, s.OrgID, s.Name, s.Location, s.Timezone, s.UpdatedAt, s.ID)
	return err
}

func DeleteStation(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM stations WHERE id = $1", id)
	return err
}

func CreateDevice(ctx context.Context, db sqlx.ExtContext, d *model.Device) error {
	query := `INSERT INTO devices (id, station_id, device_code, device_type, status, created_at, updated_at)
		VALUES (:id, :station_id, :device_code, :device_type, :status, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, d)
	return err
}

func GetDevice(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Device, error) {
	var d model.Device
	err := sqlx.GetContext(ctx, db, &d, "SELECT * FROM devices WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &d, err
}

func ListDevices(ctx context.Context, db sqlx.ExtContext, stationID uuid.UUID) ([]model.Device, error) {
	var devices []model.Device
	err := sqlx.SelectContext(ctx, db, &devices, "SELECT * FROM devices WHERE station_id = $1 ORDER BY device_code", stationID)
	return devices, err
}

func UpdateDevice(ctx context.Context, db sqlx.ExtContext, d *model.Device) error {
	query := `UPDATE devices SET station_id = $1, device_code = $2, device_type = $3, status = $4, updated_at = $5 WHERE id = $6`
	_, err := db.ExecContext(ctx, query, d.StationID, d.DeviceCode, d.DeviceType, d.Status, d.UpdatedAt, d.ID)
	return err
}

func DeleteDevice(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM devices WHERE id = $1", id)
	return err
}
