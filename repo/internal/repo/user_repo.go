package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateUser(ctx context.Context, db sqlx.ExtContext, user *model.User) error {
	query := `INSERT INTO users (id, email, password_hash, display_name, role, org_id, locked_until, created_at, updated_at)
		VALUES (:id, :email, :password_hash, :display_name, :role, :org_id, :locked_until, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, user)
	return err
}

func GetUserByID(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := sqlx.GetContext(ctx, db, &u, "SELECT * FROM users WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func GetUserByEmail(ctx context.Context, db sqlx.ExtContext, email string) (*model.User, error) {
	var u model.User
	err := sqlx.GetContext(ctx, db, &u, "SELECT * FROM users WHERE email = $1", email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func UpdateUser(ctx context.Context, db sqlx.ExtContext, user *model.User) error {
	query := `UPDATE users SET email = $1, password_hash = $2, display_name = $3, role = $4, org_id = $5, locked_until = $6, updated_at = $7 WHERE id = $8`
	_, err := db.ExecContext(ctx, query, user.Email, user.PasswordHash, user.DisplayName, user.Role, user.OrgID, user.LockedUntil, user.UpdatedAt, user.ID)
	return err
}

func DeleteUser(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

func ListUsers(ctx context.Context, db sqlx.ExtContext, limit, offset int) ([]model.User, error) {
	var users []model.User
	err := sqlx.SelectContext(ctx, db, &users, "SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return users, err
}

func UpdateRole(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, role string) error {
	_, err := db.ExecContext(ctx, "UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2", role, id)
	return err
}

func SetUserOrg(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, orgID *uuid.UUID) error {
	_, err := db.ExecContext(ctx, "UPDATE users SET org_id = $1, updated_at = NOW() WHERE id = $2", orgID, id)
	return err
}

func SetLockedUntil(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, lockedUntil time.Time) error {
	_, err := db.ExecContext(ctx, "UPDATE users SET locked_until = $1, updated_at = NOW() WHERE id = $2", lockedUntil, id)
	return err
}

func CreateSession(ctx context.Context, db sqlx.ExtContext, session *model.Session) error {
	query := `INSERT INTO sessions (id, user_id, device_id, token_hash, idle_expires_at, abs_expires_at, created_at)
		VALUES (:id, :user_id, :device_id, :token_hash, :idle_expires_at, :abs_expires_at, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, session)
	return err
}

func GetSessionByTokenHash(ctx context.Context, db sqlx.ExtContext, tokenHash string) (*model.Session, error) {
	var s model.Session
	err := sqlx.GetContext(ctx, db, &s, "SELECT * FROM sessions WHERE token_hash = $1", tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &s, err
}

func UpdateSessionIdle(ctx context.Context, db sqlx.ExtContext, sessionID uuid.UUID, newIdle time.Time) error {
	_, err := db.ExecContext(ctx, "UPDATE sessions SET idle_expires_at = $1 WHERE id = $2", newIdle, sessionID)
	return err
}

func DeleteSession(ctx context.Context, db sqlx.ExtContext, sessionID uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM sessions WHERE id = $1", sessionID)
	return err
}

func DeleteUserSessions(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	return err
}

func CreateRecoveryToken(ctx context.Context, db sqlx.ExtContext, token *model.RecoveryToken) error {
	query := `INSERT INTO recovery_tokens (id, user_id, token_hash, expires_at, used, created_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :used, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, token)
	return err
}

func GetRecoveryTokenByHash(ctx context.Context, db sqlx.ExtContext, tokenHash string) (*model.RecoveryToken, error) {
	var t model.RecoveryToken
	err := sqlx.GetContext(ctx, db, &t, "SELECT * FROM recovery_tokens WHERE token_hash = $1", tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &t, err
}

func MarkRecoveryTokenUsed(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "UPDATE recovery_tokens SET used = TRUE WHERE id = $1", id)
	return err
}

func CreateLoginAttempt(ctx context.Context, db sqlx.ExtContext, attempt *model.LoginAttempt) error {
	query := `INSERT INTO login_attempts (user_id, success, ip_address, device_id, attempted_at)
		VALUES (:user_id, :success, :ip_address, :device_id, :attempted_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, attempt)
	return err
}

func CountFailedAttempts(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := sqlx.GetContext(ctx, db, &count, "SELECT COUNT(*) FROM login_attempts WHERE user_id = $1 AND success = FALSE AND attempted_at >= $2", userID, since)
	return count, err
}

func ListPermissions(ctx context.Context, db sqlx.ExtContext) ([]model.Permission, error) {
	var perms []model.Permission
	err := sqlx.SelectContext(ctx, db, &perms, "SELECT * FROM permissions ORDER BY name")
	return perms, err
}

func GetRolePermissions(ctx context.Context, db sqlx.ExtContext, role string) ([]model.Permission, error) {
	var perms []model.Permission
	query := `SELECT p.* FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role = $1 ORDER BY p.name`
	err := sqlx.SelectContext(ctx, db, &perms, query, role)
	return perms, err
}

func GetUserPermissions(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID) ([]model.UserPermission, error) {
	var ups []model.UserPermission
	err := sqlx.SelectContext(ctx, db, &ups, "SELECT * FROM user_permissions WHERE user_id = $1", userID)
	return ups, err
}

func UpsertUserPermission(ctx context.Context, db sqlx.ExtContext, up *model.UserPermission) error {
	query := `INSERT INTO user_permissions (id, user_id, permission_id, granted, created_at)
		VALUES (:id, :user_id, :permission_id, :granted, :created_at)
		ON CONFLICT (user_id, permission_id) DO UPDATE SET granted = EXCLUDED.granted`
	_, err := sqlx.NamedExecContext(ctx, db, query, up)
	return err
}
