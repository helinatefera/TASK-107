package dto

import "github.com/google/uuid"

type CreateOrgRequest struct {
	ParentID *uuid.UUID `json:"parent_id"`
	OrgCode  string     `json:"org_code" validate:"required"`
	Name     string     `json:"name" validate:"required"`
	TaxID    *string    `json:"tax_id"`
	Address  *string    `json:"address"`
	Timezone string     `json:"timezone" validate:"required"`
}

type UpdateOrgRequest struct {
	Name     *string `json:"name"`
	TaxID    *string `json:"tax_id"`
	Address  *string `json:"address"`
	Timezone *string `json:"timezone"`
}

type OrgResponse struct {
	ID        uuid.UUID  `json:"id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	OrgCode   string     `json:"org_code"`
	Name      string     `json:"name"`
	TaxID     string     `json:"tax_id,omitempty"`
	Address   string     `json:"address,omitempty"`
	Timezone  string     `json:"timezone"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}
