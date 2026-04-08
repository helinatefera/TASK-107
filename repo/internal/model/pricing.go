package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PriceTemplate struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	OrgID     uuid.UUID  `db:"org_id" json:"org_id"`
	Name      string     `db:"name" json:"name"`
	StationID *uuid.UUID `db:"station_id" json:"station_id,omitempty"`
	DeviceID  *uuid.UUID `db:"device_id" json:"device_id,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

type PriceTemplateVersion struct {
	ID                   uuid.UUID        `db:"id" json:"id"`
	TemplateID           uuid.UUID        `db:"template_id" json:"template_id"`
	VersionNumber        int              `db:"version_number" json:"version_number"`
	EnergyRate           decimal.Decimal  `db:"energy_rate" json:"energy_rate"`
	DurationRate         decimal.Decimal  `db:"duration_rate" json:"duration_rate"`
	ServiceFee           decimal.Decimal  `db:"service_fee" json:"service_fee"`
	ApplyTax             bool             `db:"apply_tax" json:"apply_tax"`
	TaxRate              *decimal.Decimal `db:"tax_rate" json:"tax_rate,omitempty"`
	Status               string           `db:"status" json:"status"`
	EffectiveAt          *time.Time       `db:"effective_at" json:"effective_at,omitempty"`
	ClonedFromVersionID  *uuid.UUID       `db:"cloned_from_version_id" json:"cloned_from_version_id,omitempty"`
	CreatedAt            time.Time        `db:"created_at" json:"created_at"`
}

type TOURule struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	VersionID    uuid.UUID       `db:"version_id" json:"version_id"`
	DayType      string          `db:"day_type" json:"day_type"`
	StartTime    string          `db:"start_time" json:"start_time"`
	EndTime      string          `db:"end_time" json:"end_time"`
	EnergyRate   decimal.Decimal `db:"energy_rate" json:"energy_rate"`
	DurationRate decimal.Decimal `db:"duration_rate" json:"duration_rate"`
}

type OrderSnapshot struct {
	ID           uuid.UUID        `db:"id" json:"id"`
	OrderID      string           `db:"order_id" json:"order_id"`
	UserID       uuid.UUID        `db:"user_id" json:"user_id"`
	DeviceID     uuid.UUID        `db:"device_id" json:"device_id"`
	StationID    uuid.UUID        `db:"station_id" json:"station_id"`
	VersionID    uuid.UUID        `db:"version_id" json:"version_id"`
	EnergyRate   decimal.Decimal  `db:"energy_rate" json:"energy_rate"`
	DurationRate decimal.Decimal  `db:"duration_rate" json:"duration_rate"`
	ServiceFee   decimal.Decimal  `db:"service_fee" json:"service_fee"`
	TaxRate      *decimal.Decimal `db:"tax_rate" json:"tax_rate,omitempty"`
	TOUApplied   json.RawMessage  `db:"tou_applied" json:"tou_applied,omitempty"`
	EnergyKWh    decimal.Decimal  `db:"energy_kwh" json:"energy_kwh"`
	DurationMin  int              `db:"duration_min" json:"duration_min"`
	Subtotal     decimal.Decimal  `db:"subtotal" json:"subtotal"`
	TaxAmount    decimal.Decimal  `db:"tax_amount" json:"tax_amount"`
	Total        decimal.Decimal  `db:"total" json:"total"`
	OrderStart   time.Time        `db:"order_start" json:"order_start"`
	OrderEnd     time.Time        `db:"order_end" json:"order_end"`
	CreatedAt    time.Time        `db:"created_at" json:"created_at"`
}
