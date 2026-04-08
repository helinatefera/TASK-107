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

func CreateWarehouse(ctx context.Context, db sqlx.ExtContext, w *model.Warehouse) error {
	query := `INSERT INTO warehouses (id, org_id, name, address, timezone, created_at, updated_at)
		VALUES (:id, :org_id, :name, :address, :timezone, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, w)
	return err
}

func GetWarehouse(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Warehouse, error) {
	var w model.Warehouse
	err := sqlx.GetContext(ctx, db, &w, "SELECT * FROM warehouses WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &w, err
}

func ListWarehouses(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Warehouse, error) {
	var warehouses []model.Warehouse
	if orgID != nil {
		err := sqlx.SelectContext(ctx, db, &warehouses, `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id) SELECT w.* FROM warehouses w JOIN org_tree ot ON w.org_id = ot.id ORDER BY w.created_at DESC LIMIT $2 OFFSET $3`, *orgID, limit, offset)
		return warehouses, err
	}
	err := sqlx.SelectContext(ctx, db, &warehouses, "SELECT * FROM warehouses ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return warehouses, err
}

func UpdateWarehouse(ctx context.Context, db sqlx.ExtContext, w *model.Warehouse) error {
	query := `UPDATE warehouses SET org_id = $1, name = $2, address = $3, timezone = $4, updated_at = $5 WHERE id = $6`
	_, err := db.ExecContext(ctx, query, w.OrgID, w.Name, w.Address, w.Timezone, w.UpdatedAt, w.ID)
	return err
}

func DeleteWarehouse(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM warehouses WHERE id = $1", id)
	return err
}

func CreateZone(ctx context.Context, db sqlx.ExtContext, z *model.Zone) error {
	query := `INSERT INTO zones (id, warehouse_id, name, zone_type, created_at)
		VALUES (:id, :warehouse_id, :name, :zone_type, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, z)
	return err
}

func ListZones(ctx context.Context, db sqlx.ExtContext, warehouseID uuid.UUID) ([]model.Zone, error) {
	var zones []model.Zone
	err := sqlx.SelectContext(ctx, db, &zones, "SELECT * FROM zones WHERE warehouse_id = $1 ORDER BY name", warehouseID)
	return zones, err
}

func GetZone(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Zone, error) {
	var z model.Zone
	err := sqlx.GetContext(ctx, db, &z, "SELECT * FROM zones WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &z, err
}

func CreateBin(ctx context.Context, db sqlx.ExtContext, b *model.Bin) error {
	query := `INSERT INTO bins (id, zone_id, warehouse_id, bin_code, capacity, created_at)
		VALUES (:id, :zone_id, :warehouse_id, :bin_code, :capacity, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, b)
	return err
}

func ListBins(ctx context.Context, db sqlx.ExtContext, zoneID uuid.UUID) ([]model.Bin, error) {
	var bins []model.Bin
	err := sqlx.SelectContext(ctx, db, &bins, "SELECT * FROM bins WHERE zone_id = $1 ORDER BY bin_code", zoneID)
	return bins, err
}

func GetBin(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Bin, error) {
	var b model.Bin
	err := sqlx.GetContext(ctx, db, &b, "SELECT * FROM bins WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &b, err
}

func UpdateBin(ctx context.Context, db sqlx.ExtContext, b *model.Bin) error {
	query := `UPDATE bins SET zone_id = $1, warehouse_id = $2, bin_code = $3, capacity = $4 WHERE id = $5`
	_, err := db.ExecContext(ctx, query, b.ZoneID, b.WarehouseID, b.BinCode, b.Capacity, b.ID)
	return err
}

func DeleteBin(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM bins WHERE id = $1", id)
	return err
}
