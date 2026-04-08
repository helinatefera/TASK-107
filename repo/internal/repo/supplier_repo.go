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

func CreateSupplier(ctx context.Context, db sqlx.ExtContext, s *model.Supplier) error {
	query := `INSERT INTO suppliers (id, org_id, name, normalized_name, tax_id, contact_email, address, created_at, updated_at)
		VALUES (:id, :org_id, :name, :normalized_name, :tax_id, :contact_email, :address, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, s)
	return err
}

func GetSupplier(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Supplier, error) {
	var s model.Supplier
	err := sqlx.GetContext(ctx, db, &s, "SELECT * FROM suppliers WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &s, err
}

func ListSuppliers(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Supplier, error) {
	var suppliers []model.Supplier
	if orgID != nil {
		err := sqlx.SelectContext(ctx, db, &suppliers, `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id) SELECT s.* FROM suppliers s JOIN org_tree ot ON s.org_id = ot.id ORDER BY s.created_at DESC LIMIT $2 OFFSET $3`, *orgID, limit, offset)
		return suppliers, err
	}
	err := sqlx.SelectContext(ctx, db, &suppliers, "SELECT * FROM suppliers ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return suppliers, err
}

func UpdateSupplier(ctx context.Context, db sqlx.ExtContext, s *model.Supplier) error {
	query := `UPDATE suppliers SET name = $1, normalized_name = $2, tax_id = $3, contact_email = $4, address = $5, updated_at = $6, org_id = $7 WHERE id = $8`
	_, err := db.ExecContext(ctx, query, s.Name, s.NormalizedName, s.TaxID, s.ContactEmail, s.Address, s.UpdatedAt, s.OrgID, s.ID)
	return err
}

func CheckSupplierDuplicate(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, normalizedName string, taxID *string, excludeID *uuid.UUID) (bool, error) {
	var exists bool
	if taxID != nil {
		// Strict pair: match on (normalized_name AND tax_id) within the same org
		query := `SELECT EXISTS(SELECT 1 FROM suppliers WHERE ($1::uuid IS NULL OR org_id = $1) AND normalized_name = $2 AND tax_id = $3 AND ($4::uuid IS NULL OR id != $4))`
		err := sqlx.GetContext(ctx, db, &exists, query, orgID, normalizedName, *taxID, excludeID)
		return exists, err
	}
	query := `SELECT EXISTS(SELECT 1 FROM suppliers WHERE ($1::uuid IS NULL OR org_id = $1) AND normalized_name = $2 AND ($3::uuid IS NULL OR id != $3))`
	err := sqlx.GetContext(ctx, db, &exists, query, orgID, normalizedName, excludeID)
	return exists, err
}

func CreateCarrier(ctx context.Context, db sqlx.ExtContext, c *model.Carrier) error {
	query := `INSERT INTO carriers (id, org_id, name, normalized_name, tax_id, contact_email, created_at, updated_at)
		VALUES (:id, :org_id, :name, :normalized_name, :tax_id, :contact_email, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, c)
	return err
}

func GetCarrier(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Carrier, error) {
	var c model.Carrier
	err := sqlx.GetContext(ctx, db, &c, "SELECT * FROM carriers WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &c, err
}

func ListCarriers(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Carrier, error) {
	var carriers []model.Carrier
	if orgID != nil {
		err := sqlx.SelectContext(ctx, db, &carriers, `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id) SELECT c.* FROM carriers c JOIN org_tree ot ON c.org_id = ot.id ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`, *orgID, limit, offset)
		return carriers, err
	}
	err := sqlx.SelectContext(ctx, db, &carriers, "SELECT * FROM carriers ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return carriers, err
}

func UpdateCarrier(ctx context.Context, db sqlx.ExtContext, c *model.Carrier) error {
	query := `UPDATE carriers SET name = $1, normalized_name = $2, tax_id = $3, contact_email = $4, updated_at = $5, org_id = $6 WHERE id = $7`
	_, err := db.ExecContext(ctx, query, c.Name, c.NormalizedName, c.TaxID, c.ContactEmail, c.UpdatedAt, c.OrgID, c.ID)
	return err
}

func CheckCarrierDuplicate(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, normalizedName string, taxID *string, excludeID *uuid.UUID) (bool, error) {
	var exists bool
	if taxID != nil {
		// Strict pair: match on (normalized_name AND tax_id) within the same org
		query := `SELECT EXISTS(SELECT 1 FROM carriers WHERE ($1::uuid IS NULL OR org_id = $1) AND normalized_name = $2 AND tax_id = $3 AND ($4::uuid IS NULL OR id != $4))`
		err := sqlx.GetContext(ctx, db, &exists, query, orgID, normalizedName, *taxID, excludeID)
		return exists, err
	}
	query := `SELECT EXISTS(SELECT 1 FROM carriers WHERE ($1::uuid IS NULL OR org_id = $1) AND normalized_name = $2 AND ($3::uuid IS NULL OR id != $3))`
	err := sqlx.GetContext(ctx, db, &exists, query, orgID, normalizedName, excludeID)
	return exists, err
}
