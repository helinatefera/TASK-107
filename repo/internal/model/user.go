package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID    `db:"id" json:"id"`
	Email        string       `db:"email" json:"email"`
	PasswordHash string       `db:"password_hash" json:"-"`
	DisplayName  string       `db:"display_name" json:"display_name"`
	Role         string       `db:"role" json:"role"`
	OrgID        *uuid.UUID   `db:"org_id" json:"org_id,omitempty"`
	LockedUntil  sql.NullTime `db:"locked_until" json:"-"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
}

type Session struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	DeviceID      string    `db:"device_id" json:"device_id"`
	TokenHash     string    `db:"token_hash" json:"-"`
	IdleExpiresAt time.Time `db:"idle_expires_at" json:"idle_expires_at"`
	AbsExpiresAt  time.Time `db:"abs_expires_at" json:"abs_expires_at"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type RecoveryToken struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	TokenHash string    `db:"token_hash" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type LoginAttempt struct {
	ID          int64          `db:"id"`
	UserID      uuid.UUID      `db:"user_id"`
	Success     bool           `db:"success"`
	IPAddress   sql.NullString `db:"ip_address"`
	DeviceID    sql.NullString `db:"device_id"`
	AttemptedAt time.Time      `db:"attempted_at"`
}

type Permission struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type RolePermission struct {
	Role         string    `db:"role" json:"role"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}

type UserPermission struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
	Granted      bool      `db:"granted" json:"granted"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
