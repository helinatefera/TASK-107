package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateCarouselRequest struct {
	OrgID      uuid.UUID `json:"org_id" validate:"required"`
	Title      string    `json:"title" validate:"required"`
	ImageURL   *string   `json:"image_url"`
	LinkURL    *string   `json:"link_url"`
	Priority   int       `json:"priority"`
	TargetRole *string   `json:"target_role" validate:"omitempty,oneof=guest user merchant administrator"`
	StartTime  time.Time `json:"start_time" validate:"required"`
	EndTime    time.Time `json:"end_time" validate:"required"`
}

type UpdateCarouselRequest struct {
	Title      *string    `json:"title"`
	ImageURL   *string    `json:"image_url"`
	LinkURL    *string    `json:"link_url"`
	Priority   *int       `json:"priority"`
	TargetRole *string    `json:"target_role" validate:"omitempty,oneof=guest user merchant administrator"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Active     *bool      `json:"active"`
}

type CreateCampaignRequest struct {
	OrgID      uuid.UUID       `json:"org_id" validate:"required"`
	Name       string          `json:"name" validate:"required"`
	Content    json.RawMessage `json:"content" validate:"required"`
	Priority   int             `json:"priority"`
	TargetRole *string         `json:"target_role" validate:"omitempty,oneof=guest user merchant administrator"`
	StartTime  time.Time       `json:"start_time" validate:"required"`
	EndTime    time.Time       `json:"end_time" validate:"required"`
}

type UpdateCampaignRequest struct {
	Name       *string          `json:"name"`
	Content    *json.RawMessage `json:"content"`
	Priority   *int             `json:"priority"`
	TargetRole *string          `json:"target_role" validate:"omitempty,oneof=guest user merchant administrator"`
	StartTime  *time.Time       `json:"start_time"`
	EndTime    *time.Time       `json:"end_time"`
	Active     *bool            `json:"active"`
}

type CreateRankingRequest struct {
	OrgID      uuid.UUID `json:"org_id" validate:"required"`
	EntityType string    `json:"entity_type" validate:"required"`
	EntityID   uuid.UUID `json:"entity_id" validate:"required"`
	Score      int       `json:"score"`
	Priority   int       `json:"priority"`
	TargetRole *string   `json:"target_role" validate:"omitempty,oneof=guest user merchant administrator"`
	StartTime  time.Time `json:"start_time" validate:"required"`
	EndTime    time.Time `json:"end_time" validate:"required"`
}

type UpdateRankingRequest struct {
	Score      *int       `json:"score"`
	Priority   *int       `json:"priority"`
	TargetRole *string    `json:"target_role" validate:"omitempty,oneof=guest user merchant administrator"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Active     *bool      `json:"active"`
}
