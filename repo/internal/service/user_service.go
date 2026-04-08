package service

import (
	"context"
	"time"

	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func GetUser(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID) (*model.User, error) {
	return repo.GetUserByID(ctx, db, userID)
}

func ListUsers(ctx context.Context, db sqlx.ExtContext, limit, offset int) ([]model.User, error) {
	return repo.ListUsers(ctx, db, limit, offset)
}

func UpdateUser(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, req *dto.UpdateUserRequest) (*model.User, error) {
	user, err := repo.GetUserByID(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Email != nil {
		user.Email = *req.Email
	}

	user.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateUser(ctx, db, user); err != nil {
		return nil, err
	}
	return user, nil
}

func DeleteUser(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID) error {
	return repo.DeleteUser(ctx, db, userID)
}

func UpdateUserRole(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, role string) error {
	return repo.UpdateRole(ctx, db, userID, role)
}

func SetUserOrg(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, orgID *uuid.UUID) error {
	return repo.SetUserOrg(ctx, db, userID, orgID)
}

func GetEffectivePermissions(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, role string) ([]dto.Permission, error) {
	rolePerms, err := repo.GetRolePermissions(ctx, db, role)
	if err != nil {
		return nil, err
	}

	userPerms, err := repo.GetUserPermissions(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	// Build a map of user-level overrides keyed by permission ID
	overrides := make(map[uuid.UUID]bool, len(userPerms))
	for _, up := range userPerms {
		overrides[up.PermissionID] = up.Granted
	}

	// Start with role permissions (all granted by default)
	effective := make(map[uuid.UUID]dto.Permission, len(rolePerms))
	for _, rp := range rolePerms {
		effective[rp.ID] = dto.Permission{
			ID:      rp.ID,
			Name:    rp.Name,
			Granted: true,
		}
	}

	// Build a name lookup from all permissions for resolving user-only grants
	allPerms, _ := repo.ListPermissions(ctx, db)
	permNames := make(map[uuid.UUID]string, len(allPerms))
	for _, p := range allPerms {
		permNames[p.ID] = p.Name
	}

	// Apply user overrides: may revoke role permissions or grant additional ones
	for _, up := range userPerms {
		if up.Granted {
			if _, exists := effective[up.PermissionID]; !exists {
				effective[up.PermissionID] = dto.Permission{
					ID:      up.PermissionID,
					Name:    permNames[up.PermissionID],
					Granted: true,
				}
			}
		} else {
			// Revoke: if the role granted it, mark as not granted
			if p, exists := effective[up.PermissionID]; exists {
				p.Granted = false
				effective[up.PermissionID] = p
			}
		}
	}

	result := make([]dto.Permission, 0, len(effective))
	for _, p := range effective {
		result = append(result, p)
	}
	return result, nil
}

func UpdateUserPermissions(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, grants []dto.PermissionGrant) error {
	now := time.Now().UTC()
	for _, g := range grants {
		up := &model.UserPermission{
			ID:           uuid.New(),
			UserID:       userID,
			PermissionID: g.PermissionID,
			Granted:      g.Granted,
			CreatedAt:    now,
		}
		if err := repo.UpsertUserPermission(ctx, db, up); err != nil {
			return err
		}
	}
	return nil
}
