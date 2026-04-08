package dto

import "github.com/google/uuid"

type CreateWarehouseRequest struct {
	OrgID    uuid.UUID `json:"org_id" validate:"required"`
	Name     string    `json:"name" validate:"required"`
	Address  *string   `json:"address"`
	Timezone string    `json:"timezone" validate:"required"`
}

type UpdateWarehouseRequest struct {
	Name     *string `json:"name"`
	Address  *string `json:"address"`
	Timezone *string `json:"timezone"`
}

type CreateZoneRequest struct {
	Name     string  `json:"name" validate:"required"`
	ZoneType *string `json:"zone_type"`
}

type CreateBinRequest struct {
	BinCode  string `json:"bin_code" validate:"required"`
	Capacity *int   `json:"capacity"`
}

type UpdateBinRequest struct {
	BinCode  *string `json:"bin_code"`
	Capacity *int    `json:"capacity"`
}
