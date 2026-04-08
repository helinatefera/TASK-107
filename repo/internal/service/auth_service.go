package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
	"unicode"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/db"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return err != nil && errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func ValidatePassword(pw string) error {
	if len(pw) < 12 {
		return apperror.ErrPasswordTooShort
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, ch := range pw {
		switch {
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	classes := 0
	if hasLower {
		classes++
	}
	if hasUpper {
		classes++
	}
	if hasDigit {
		classes++
	}
	if hasSpecial {
		classes++
	}

	if classes < 3 {
		return apperror.ErrPasswordComplexity
	}
	return nil
}

func Register(ctx context.Context, database *sqlx.DB, req *dto.RegisterRequest, cfg *config.Config) (*model.User, error) {
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), cfg.BcryptCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repo.CreateUser(ctx, database, user); err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}
	return user, nil
}

func Login(ctx context.Context, database *sqlx.DB, req *dto.LoginRequest, cfg *config.Config) (string, *model.Session, error) {
	user, err := repo.GetUserByEmail(ctx, database, req.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return "", nil, apperror.ErrUnauthorized
		}
		return "", nil, err
	}

	now := time.Now().UTC()

	// Check lockout
	if user.LockedUntil.Valid && user.LockedUntil.Time.After(now) {
		return "", nil, apperror.ErrAccountLocked
	}

	windowStart := now.Add(-cfg.LockoutWindow)

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Record failed attempt before checking threshold
		_ = repo.CreateLoginAttempt(ctx, database, &model.LoginAttempt{
			UserID:      user.ID,
			Success:     false,
			AttemptedAt: now,
		})
		// Re-count after recording this attempt
		failedCount, countErr := repo.CountFailedAttempts(ctx, database, user.ID, windowStart)
		if countErr == nil && failedCount >= cfg.LockoutMax {
			lockedUntil := now.Add(cfg.LockoutDuration)
			_ = repo.SetLockedUntil(ctx, database, user.ID, lockedUntil)
			return "", nil, apperror.ErrAccountLocked
		}
		return "", nil, apperror.ErrUnauthorized
	}

	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, err
	}
	plainToken := hex.EncodeToString(tokenBytes)

	tokenHashBytes := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	session := &model.Session{
		ID:            uuid.New(),
		UserID:        user.ID,
		DeviceID:      req.DeviceID,
		TokenHash:     tokenHash,
		IdleExpiresAt: now.Add(cfg.SessionIdleTime),
		AbsExpiresAt:  now.Add(cfg.SessionAbsTime),
		CreatedAt:     now,
	}

	if err := repo.CreateSession(ctx, database, session); err != nil {
		return "", nil, err
	}

	// Record successful attempt
	_ = repo.CreateLoginAttempt(ctx, database, &model.LoginAttempt{
		UserID:      user.ID,
		Success:     true,
		AttemptedAt: now,
	})

	return plainToken, session, nil
}

func Logout(ctx context.Context, database *sqlx.DB, sessionID uuid.UUID) error {
	return repo.DeleteSession(ctx, database, sessionID)
}

func RefreshSession(ctx context.Context, database *sqlx.DB, session *model.Session, deviceID string, cfg *config.Config) error {
	if session.DeviceID != deviceID {
		return apperror.ErrDeviceMismatch
	}

	now := time.Now().UTC()

	// Enforce idle timeout — a session that has already idle-expired cannot be refreshed
	if now.After(session.IdleExpiresAt) {
		return apperror.ErrSessionExpired
	}

	if session.AbsExpiresAt.Before(now) {
		return apperror.ErrSessionExpired
	}

	newIdle := now.Add(cfg.SessionIdleTime)
	return repo.UpdateSessionIdle(ctx, database, session.ID, newIdle)
}

func RequestRecovery(ctx context.Context, database *sqlx.DB, email string, cfg *config.Config) (string, error) {
	user, err := repo.GetUserByEmail(ctx, database, email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			// Return a fake token so the response is indistinguishable from a real one
			fakeBytes := make([]byte, 32)
			rand.Read(fakeBytes)
			return hex.EncodeToString(fakeBytes), nil
		}
		return "", err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	plainToken := hex.EncodeToString(tokenBytes)

	tokenHashBytes := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	now := time.Now().UTC()
	rt := &model.RecoveryToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(cfg.RecoveryTTL),
		Used:      false,
		CreatedAt: now,
	}

	if err := repo.CreateRecoveryToken(ctx, database, rt); err != nil {
		return "", err
	}

	return plainToken, nil
}

func ResetPassword(ctx context.Context, database *sqlx.DB, req *dto.ResetPasswordRequest, cfg *config.Config) error {
	tokenHashBytes := sha256.Sum256([]byte(req.Token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	rt, err := repo.GetRecoveryTokenByHash(ctx, database, tokenHash)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrInvalidToken
		}
		return err
	}

	if rt.Used {
		return apperror.ErrInvalidToken
	}

	now := time.Now().UTC()
	if rt.ExpiresAt.Before(now) {
		return apperror.ErrInvalidToken
	}

	if err := ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), cfg.BcryptCost)
	if err != nil {
		return err
	}

	return db.WithTx(ctx, database, func(tx *sqlx.Tx) error {
		user, err := repo.GetUserByID(ctx, tx, rt.UserID)
		if err != nil {
			return err
		}

		user.PasswordHash = string(hash)
		user.UpdatedAt = now
		if err := repo.UpdateUser(ctx, tx, user); err != nil {
			return err
		}

		if err := repo.MarkRecoveryTokenUsed(ctx, tx, rt.ID); err != nil {
			return err
		}

		return repo.DeleteUserSessions(ctx, tx, user.ID)
	})
}
