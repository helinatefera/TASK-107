package model

import (
	"time"

	"github.com/google/uuid"
)

type Station struct {
	ID        uuid.UUID `db:"id" json:"id"`
	OrgID     uuid.UUID `db:"org_id" json:"org_id"`
	Name      string    `db:"name" json:"name"`
	Location  *string   `db:"location" json:"location,omitempty"`
	Timezone  string    `db:"timezone" json:"timezone"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Device struct {
	ID         uuid.UUID `db:"id" json:"id"`
	StationID  uuid.UUID `db:"station_id" json:"station_id"`
	DeviceCode string    `db:"device_code" json:"device_code"`
	DeviceType *string   `db:"device_type" json:"device_type,omitempty"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
