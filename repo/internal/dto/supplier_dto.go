package dto

import "github.com/google/uuid"

type CreateSupplierRequest struct {
	OrgID        uuid.UUID `json:"org_id" validate:"required"`
	Name         string    `json:"name" validate:"required"`
	TaxID        *string   `json:"tax_id"`
	ContactEmail *string   `json:"contact_email" validate:"omitempty,email"`
	Address      *string   `json:"address"`
}

type UpdateSupplierRequest struct {
	Name         *string `json:"name"`
	TaxID        *string `json:"tax_id"`
	ContactEmail *string `json:"contact_email" validate:"omitempty,email"`
	Address      *string `json:"address"`
}

type CreateCarrierRequest struct {
	OrgID        uuid.UUID `json:"org_id" validate:"required"`
	Name         string    `json:"name" validate:"required"`
	TaxID        *string   `json:"tax_id"`
	ContactEmail *string   `json:"contact_email" validate:"omitempty,email"`
}

type UpdateCarrierRequest struct {
	Name         *string `json:"name"`
	TaxID        *string `json:"tax_id"`
	ContactEmail *string `json:"contact_email" validate:"omitempty,email"`
}
