package model

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID        uuid.UUID `db:"id" json:"id"`
	OrgID     uuid.UUID `db:"org_id" json:"org_id"`
	Name      string    `db:"name" json:"name"`
	Address   *string   `db:"address" json:"address,omitempty"`
	Timezone  string    `db:"timezone" json:"timezone"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Zone struct {
	ID          uuid.UUID `db:"id" json:"id"`
	WarehouseID uuid.UUID `db:"warehouse_id" json:"warehouse_id"`
	Name        string    `db:"name" json:"name"`
	ZoneType    *string   `db:"zone_type" json:"zone_type,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Bin struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ZoneID      uuid.UUID `db:"zone_id" json:"zone_id"`
	WarehouseID uuid.UUID `db:"warehouse_id" json:"warehouse_id"`
	BinCode     string    `db:"bin_code" json:"bin_code"`
	Capacity    *int      `db:"capacity" json:"capacity,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
