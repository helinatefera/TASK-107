package model

import (
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	OrgID          *uuid.UUID `db:"org_id" json:"org_id,omitempty"`
	Name           string     `db:"name" json:"name"`
	NormalizedName string     `db:"normalized_name" json:"-"`
	TaxID          *string    `db:"tax_id" json:"tax_id,omitempty"`
	ContactEmail   *string    `db:"contact_email" json:"contact_email,omitempty"`
	Address        *string    `db:"address" json:"address,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

type Carrier struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	OrgID          *uuid.UUID `db:"org_id" json:"org_id,omitempty"`
	Name           string     `db:"name" json:"name"`
	NormalizedName string     `db:"normalized_name" json:"-"`
	TaxID          *string    `db:"tax_id" json:"tax_id,omitempty"`
	ContactEmail   *string    `db:"contact_email" json:"contact_email,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}
