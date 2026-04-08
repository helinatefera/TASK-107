package model

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         int64            `db:"id" json:"id"`
	UserID     *string          `db:"user_id" json:"user_id,omitempty"`
	Action     string           `db:"action" json:"action"`
	EntityType string           `db:"entity_type" json:"entity_type"`
	EntityID   string           `db:"entity_id" json:"entity_id"`
	OldValue   *json.RawMessage `db:"old_value" json:"old_value,omitempty"`
	NewValue   *json.RawMessage `db:"new_value" json:"new_value,omitempty"`
	IPAddress  *string          `db:"ip_address" json:"ip_address,omitempty"`
	RequestID  *string          `db:"request_id" json:"request_id,omitempty"`
	CreatedAt  time.Time        `db:"created_at" json:"created_at"`
}

type AppConfig struct {
	Key       string    `db:"key" json:"key"`
	Value     string    `db:"value" json:"value"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type RequestMetric struct {
	ID         int64     `db:"id" json:"id"`
	Method     string    `db:"method" json:"method"`
	Path       string    `db:"path" json:"path"`
	StatusCode int       `db:"status_code" json:"status_code"`
	LatencyMs  int       `db:"latency_ms" json:"latency_ms"`
	RecordedAt time.Time `db:"recorded_at" json:"recorded_at"`
}
