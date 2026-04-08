package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateNotificationTemplateRequest struct {
	Code           string `json:"code" validate:"required"`
	TitleTmpl      string `json:"title_tmpl" validate:"required"`
	BodyTmpl       string `json:"body_tmpl" validate:"required"`
	DefaultEnabled *bool  `json:"default_enabled"`
}

type UpdateSubscriptionRequest struct {
	OptedOut bool `json:"opted_out"`
}

type SendNotificationRequest struct {
	UserID     uuid.UUID       `json:"user_id" validate:"required"`
	TemplateID uuid.UUID       `json:"template_id" validate:"required"`
	Params     json.RawMessage `json:"params"`
}
