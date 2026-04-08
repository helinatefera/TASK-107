package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateCategoryRequest struct {
	Name     string     `json:"name" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type CreateItemRequest struct {
	SKU         string     `json:"sku" validate:"required"`
	ItemName    string     `json:"item_name" validate:"required"`
	CategoryID  *uuid.UUID `json:"category_id"`
	BaseUnitID  uuid.UUID  `json:"base_unit_id" validate:"required"`
	Description *string    `json:"description"`
}

type UpdateItemRequest struct {
	ItemName    *string    `json:"item_name"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Description *string    `json:"description"`
}

type CreateUnitRequest struct {
	Name   string `json:"name" validate:"required"`
	Symbol string `json:"symbol" validate:"required"`
}

type CreateConversionRequest struct {
	FromUnitID uuid.UUID       `json:"from_unit_id" validate:"required"`
	ToUnitID   uuid.UUID       `json:"to_unit_id" validate:"required"`
	Factor     decimal.Decimal `json:"factor" validate:"required"`
}
