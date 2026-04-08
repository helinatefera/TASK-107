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

func CreateStation(ctx context.Context, db sqlx.ExtContext, req *dto.CreateStationRequest) (*model.Station, error) {
	now := time.Now().UTC()
	s := &model.Station{
		ID:        uuid.New(),
		OrgID:     req.OrgID,
		Name:      req.Name,
		Location:  req.Location,
		Timezone:  req.Timezone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateStation(ctx, db, s); err != nil {
		return nil, err
	}
	return s, nil
}

func GetStation(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Station, error) {
	return repo.GetStation(ctx, db, id)
}

func ListStations(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Station, error) {
	return repo.ListStations(ctx, db, orgID, limit, offset)
}

func UpdateStation(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateStationRequest) (*model.Station, error) {
	s, err := repo.GetStation(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.Location != nil {
		s.Location = req.Location
	}
	if req.Timezone != nil {
		s.Timezone = *req.Timezone
	}

	s.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateStation(ctx, db, s); err != nil {
		return nil, err
	}
	return s, nil
}

func DeleteStation(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteStation(ctx, db, id)
}

func CreateDevice(ctx context.Context, db sqlx.ExtContext, stationID uuid.UUID, req *dto.CreateDeviceRequest) (*model.Device, error) {
	now := time.Now().UTC()
	d := &model.Device{
		ID:         uuid.New(),
		StationID:  stationID,
		DeviceCode: req.DeviceCode,
		DeviceType: req.DeviceType,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.CreateDevice(ctx, db, d); err != nil {
		return nil, err
	}
	return d, nil
}

func GetDevice(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Device, error) {
	return repo.GetDevice(ctx, db, id)
}

func ListDevices(ctx context.Context, db sqlx.ExtContext, stationID uuid.UUID) ([]model.Device, error) {
	return repo.ListDevices(ctx, db, stationID)
}

func UpdateDevice(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateDeviceRequest) (*model.Device, error) {
	d, err := repo.GetDevice(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.DeviceType != nil {
		d.DeviceType = req.DeviceType
	}
	if req.Status != nil {
		d.Status = *req.Status
	}

	d.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateDevice(ctx, db, d); err != nil {
		return nil, err
	}
	return d, nil
}

func DeleteDevice(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteDevice(ctx, db, id)
}
