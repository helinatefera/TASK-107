package model

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	ParentID  *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	OrgCode   string     `db:"org_code" json:"org_code"`
	Name      string     `db:"name" json:"name"`
	TaxID     *string    `db:"tax_id" json:"tax_id,omitempty"`
	Address   *string    `db:"address" json:"address,omitempty"`
	Timezone  string     `db:"timezone" json:"timezone"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}
