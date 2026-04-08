package service

import (
	"context"
	"time"

	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateWarehouse(ctx context.Context, db sqlx.ExtContext, req *dto.CreateWarehouseRequest) (*model.Warehouse, error) {
	now := time.Now().UTC()
	w := &model.Warehouse{
		ID:        uuid.New(),
		OrgID:     req.OrgID,
		Name:      req.Name,
		Address:   req.Address,
		Timezone:  req.Timezone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateWarehouse(ctx, db, w); err != nil {
		return nil, err
	}
	return w, nil
}

func GetWarehouse(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Warehouse, error) {
	return repo.GetWarehouse(ctx, db, id)
}

func ListWarehouses(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Warehouse, error) {
	return repo.ListWarehouses(ctx, db, orgID, limit, offset)
}

func UpdateWarehouse(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateWarehouseRequest) (*model.Warehouse, error) {
	w, err := repo.GetWarehouse(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		w.Name = *req.Name
	}
	if req.Address != nil {
		w.Address = req.Address
	}
	if req.Timezone != nil {
		w.Timezone = *req.Timezone
	}

	w.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateWarehouse(ctx, db, w); err != nil {
		return nil, err
	}
	return w, nil
}

func DeleteWarehouse(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteWarehouse(ctx, db, id)
}

func CreateZone(ctx context.Context, db sqlx.ExtContext, warehouseID uuid.UUID, req *dto.CreateZoneRequest) (*model.Zone, error) {
	z := &model.Zone{
		ID:          uuid.New(),
		WarehouseID: warehouseID,
		Name:        req.Name,
		ZoneType:    req.ZoneType,
		CreatedAt:   time.Now().UTC(),
	}
	if err := repo.CreateZone(ctx, db, z); err != nil {
		return nil, err
	}
	return z, nil
}

func GetZone(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Zone, error) {
	return repo.GetZone(ctx, db, id)
}

func ListZones(ctx context.Context, db sqlx.ExtContext, warehouseID uuid.UUID) ([]model.Zone, error) {
	return repo.ListZones(ctx, db, warehouseID)
}

func CreateBin(ctx context.Context, db sqlx.ExtContext, zoneID uuid.UUID, warehouseID uuid.UUID, req *dto.CreateBinRequest) (*model.Bin, error) {
	b := &model.Bin{
		ID:          uuid.New(),
		ZoneID:      zoneID,
		WarehouseID: warehouseID,
		BinCode:     req.BinCode,
		Capacity:    req.Capacity,
		CreatedAt:   time.Now().UTC(),
	}
	if err := repo.CreateBin(ctx, db, b); err != nil {
		return nil, err
	}
	return b, nil
}

func GetBin(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Bin, error) {
	return repo.GetBin(ctx, db, id)
}

func ListBins(ctx context.Context, db sqlx.ExtContext, zoneID uuid.UUID) ([]model.Bin, error) {
	return repo.ListBins(ctx, db, zoneID)
}

func UpdateBin(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateBinRequest) (*model.Bin, error) {
	b, err := repo.GetBin(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.BinCode != nil {
		b.BinCode = *req.BinCode
	}
	if req.Capacity != nil {
		b.Capacity = req.Capacity
	}

	if err := repo.UpdateBin(ctx, db, b); err != nil {
		return nil, err
	}
	return b, nil
}

func DeleteBin(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteBin(ctx, db, id)
}
