package dto

import "github.com/google/uuid"

type CreateStationRequest struct {
	OrgID    uuid.UUID `json:"org_id" validate:"required"`
	Name     string    `json:"name" validate:"required"`
	Location *string   `json:"location"`
	Timezone string    `json:"timezone" validate:"required"`
}

type UpdateStationRequest struct {
	Name     *string `json:"name"`
	Location *string `json:"location"`
	Timezone *string `json:"timezone"`
}

type CreateDeviceRequest struct {
	DeviceCode string  `json:"device_code" validate:"required"`
	DeviceType *string `json:"device_type"`
}

type UpdateDeviceRequest struct {
	DeviceType *string `json:"device_type"`
	Status     *string `json:"status" validate:"omitempty,oneof=active inactive maintenance"`
}
