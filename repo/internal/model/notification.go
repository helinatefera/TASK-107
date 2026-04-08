package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationTemplate struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Code           string    `db:"code" json:"code"`
	TitleTmpl      string    `db:"title_tmpl" json:"title_tmpl"`
	BodyTmpl       string    `db:"body_tmpl" json:"body_tmpl"`
	DefaultEnabled bool      `db:"default_enabled" json:"default_enabled"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type NotificationSubscription struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	TemplateID uuid.UUID `db:"template_id" json:"template_id"`
	OptedOut   bool      `db:"opted_out" json:"opted_out"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type NotificationJob struct {
	ID             int64           `db:"id" json:"id"`
	UserID         uuid.UUID       `db:"user_id" json:"user_id"`
	TemplateID     uuid.UUID       `db:"template_id" json:"template_id"`
	Params         *json.RawMessage `db:"params" json:"params,omitempty"`
	Status         string          `db:"status" json:"status"`
	SuppressReason sql.NullString  `db:"suppress_reason" json:"suppress_reason,omitempty"`
	ScheduledAt    time.Time       `db:"scheduled_at" json:"scheduled_at"`
	ProcessedAt    sql.NullTime    `db:"processed_at" json:"processed_at,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

type Message struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	TemplateID uuid.UUID `db:"template_id" json:"template_id"`
	Title      string    `db:"title" json:"title"`
	Body       string    `db:"body" json:"body"`
	Read       bool      `db:"read" json:"read"`
	Dismissed  bool      `db:"dismissed" json:"dismissed"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type DeliveryStats struct {
	ID         int64     `db:"id" json:"id"`
	TemplateID uuid.UUID `db:"template_id" json:"template_id"`
	Date       time.Time `db:"date" json:"date"`
	Generated  int       `db:"generated" json:"generated"`
	Delivered  int       `db:"delivered" json:"delivered"`
	Opened     int       `db:"opened" json:"opened"`
	Dismissed  int       `db:"dismissed" json:"dismissed"`
}
