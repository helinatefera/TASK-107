package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CarouselSlot struct {
	ID         uuid.UUID `db:"id" json:"id"`
	OrgID      uuid.UUID `db:"org_id" json:"org_id"`
	Title      string    `db:"title" json:"title"`
	ImageURL   *string   `db:"image_url" json:"image_url,omitempty"`
	LinkURL    *string   `db:"link_url" json:"link_url,omitempty"`
	Priority   int       `db:"priority" json:"priority"`
	TargetRole *string   `db:"target_role" json:"target_role,omitempty"`
	StartTime  time.Time `db:"start_time" json:"start_time"`
	EndTime    time.Time `db:"end_time" json:"end_time"`
	Active     bool      `db:"active" json:"active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type CampaignPlacement struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	OrgID      uuid.UUID       `db:"org_id" json:"org_id"`
	Name       string          `db:"name" json:"name"`
	Content    json.RawMessage `db:"content" json:"content"`
	Priority   int             `db:"priority" json:"priority"`
	TargetRole *string         `db:"target_role" json:"target_role,omitempty"`
	StartTime  time.Time       `db:"start_time" json:"start_time"`
	EndTime    time.Time       `db:"end_time" json:"end_time"`
	Active     bool            `db:"active" json:"active"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at" json:"updated_at"`
}

type HotRanking struct {
	ID         uuid.UUID `db:"id" json:"id"`
	OrgID      uuid.UUID `db:"org_id" json:"org_id"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   uuid.UUID `db:"entity_id" json:"entity_id"`
	Score      int       `db:"score" json:"score"`
	Priority   int       `db:"priority" json:"priority"`
	TargetRole *string   `db:"target_role" json:"target_role,omitempty"`
	StartTime  time.Time `db:"start_time" json:"start_time"`
	EndTime    time.Time `db:"end_time" json:"end_time"`
	Active     bool      `db:"active" json:"active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
