package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateOrg(ctx context.Context, db sqlx.ExtContext, org *model.Organization) error {
	query := `INSERT INTO organizations (id, parent_id, org_code, name, tax_id, address, timezone, created_at, updated_at)
		VALUES (:id, :parent_id, :org_code, :name, :tax_id, :address, :timezone, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, org)
	return err
}

func GetOrgByID(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	err := sqlx.GetContext(ctx, db, &org, "SELECT * FROM organizations WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &org, err
}

func ListOrgs(ctx context.Context, db sqlx.ExtContext, limit, offset int) ([]model.Organization, error) {
	var orgs []model.Organization
	err := sqlx.SelectContext(ctx, db, &orgs, "SELECT * FROM organizations ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return orgs, err
}

func ListOrgsByOrgID(ctx context.Context, db sqlx.ExtContext, orgID uuid.UUID, limit, offset int) ([]model.Organization, error) {
	var orgs []model.Organization
	query := `WITH RECURSIVE org_tree AS (
		SELECT id FROM organizations WHERE id = $1
		UNION ALL
		SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id
	)
	SELECT organizations.* FROM organizations JOIN org_tree ON organizations.id = org_tree.id ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := sqlx.SelectContext(ctx, db, &orgs, query, orgID, limit, offset)
	return orgs, err
}

func UpdateOrg(ctx context.Context, db sqlx.ExtContext, org *model.Organization) error {
	query := `UPDATE organizations SET parent_id = $1, org_code = $2, name = $3, tax_id = $4, address = $5, timezone = $6, updated_at = $7 WHERE id = $8`
	_, err := db.ExecContext(ctx, query, org.ParentID, org.OrgCode, org.Name, org.TaxID, org.Address, org.Timezone, org.UpdatedAt, org.ID)
	return err
}

func DeleteOrg(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM organizations WHERE id = $1", id)
	return err
}

// IsOrgAccessible returns true if resourceOrgID equals callerOrgID or is a
// descendant of callerOrgID at any depth in the org hierarchy.
func IsOrgAccessible(ctx context.Context, db sqlx.ExtContext, callerOrgID, resourceOrgID uuid.UUID) (bool, error) {
	if callerOrgID == resourceOrgID {
		return true, nil
	}
	var count int
	query := `WITH RECURSIVE org_tree AS (
		SELECT id FROM organizations WHERE id = $1
		UNION ALL
		SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id
	)
	SELECT COUNT(*) FROM org_tree WHERE id = $2`
	err := sqlx.GetContext(ctx, db, &count, query, callerOrgID, resourceOrgID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
