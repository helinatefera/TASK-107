package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Category struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	ParentID  *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

type UnitOfMeasure struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Symbol    string    `db:"symbol" json:"symbol"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type UnitConversion struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	FromUnitID uuid.UUID       `db:"from_unit_id" json:"from_unit_id"`
	ToUnitID   uuid.UUID       `db:"to_unit_id" json:"to_unit_id"`
	Factor     decimal.Decimal `db:"factor" json:"factor"`
}

type Item struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	SKU         string     `db:"sku" json:"sku"`
	ItemName    string     `db:"item_name" json:"item_name"`
	CategoryID  *uuid.UUID `db:"category_id" json:"category_id,omitempty"`
	BaseUnitID  uuid.UUID  `db:"base_unit_id" json:"base_unit_id"`
	Description *string    `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}
