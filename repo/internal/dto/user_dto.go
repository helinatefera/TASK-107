package dto

import "github.com/google/uuid"

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email" validate:"omitempty,email"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=guest user merchant administrator"`
}

type UpdatePermissionsRequest struct {
	Permissions []PermissionGrant `json:"permissions" validate:"required"`
}

type PermissionGrant struct {
	PermissionID uuid.UUID `json:"permission_id" validate:"required"`
	Granted      bool      `json:"granted"`
}

type UserResponse struct {
	ID          uuid.UUID    `json:"id"`
	Email       string       `json:"email"`
	DisplayName string       `json:"display_name"`
	Role        string       `json:"role"`
	OrgID       *uuid.UUID   `json:"org_id,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

type Permission struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Granted bool      `json:"granted"`
}
