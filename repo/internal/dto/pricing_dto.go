package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatePriceTemplateRequest struct {
	Name      string     `json:"name" validate:"required"`
	StationID *uuid.UUID `json:"station_id"`
	DeviceID  *uuid.UUID `json:"device_id"`
}

type CreateVersionRequest struct {
	EnergyRate   decimal.Decimal `json:"energy_rate" validate:"required"`
	DurationRate decimal.Decimal `json:"duration_rate" validate:"required"`
	ServiceFee   decimal.Decimal `json:"service_fee" validate:"required"`
	ApplyTax     bool            `json:"apply_tax"`
}

type CreateTOURuleRequest struct {
	DayType      string          `json:"day_type" validate:"required,oneof=weekday weekend holiday"`
	StartTime    string          `json:"start_time" validate:"required"`
	EndTime      string          `json:"end_time" validate:"required"`
	EnergyRate   decimal.Decimal `json:"energy_rate" validate:"required"`
	DurationRate decimal.Decimal `json:"duration_rate" validate:"required"`
}

// ValidateRatesNonNegative checks that all rate fields are >= 0.
// Called explicitly since go-playground/validator doesn't natively support decimal.Decimal gte.
func (r *CreateVersionRequest) ValidateRatesNonNegative() error {
	if r.EnergyRate.IsNegative() || r.DurationRate.IsNegative() || r.ServiceFee.IsNegative() {
		return fmt.Errorf("rates and fees must not be negative")
	}
	return nil
}

// ValidateRatesNonNegative checks that all rate fields are >= 0.
func (r *CreateTOURuleRequest) ValidateRatesNonNegative() error {
	if r.EnergyRate.IsNegative() || r.DurationRate.IsNegative() {
		return fmt.Errorf("rates must not be negative")
	}
	return nil
}

type ActivateVersionRequest struct {
	EffectiveAt *time.Time `json:"effective_at"`
}

type CreateOrderRequest struct {
	OrderID   string          `json:"order_id" validate:"required"`
	DeviceID  uuid.UUID       `json:"device_id" validate:"required"`
	EnergyKWh decimal.Decimal `json:"energy_kwh" validate:"required"`
	StartTime time.Time       `json:"start_time" validate:"required"`
	EndTime   time.Time       `json:"end_time" validate:"required"`
}
