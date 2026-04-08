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

func CreateCategory(ctx context.Context, db sqlx.ExtContext, c *model.Category) error {
	query := `INSERT INTO categories (id, name, parent_id, created_at)
		VALUES (:id, :name, :parent_id, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, c)
	return err
}

func ListCategories(ctx context.Context, db sqlx.ExtContext) ([]model.Category, error) {
	var cats []model.Category
	err := sqlx.SelectContext(ctx, db, &cats, "SELECT * FROM categories ORDER BY name")
	return cats, err
}

func CreateItem(ctx context.Context, db sqlx.ExtContext, item *model.Item) error {
	query := `INSERT INTO items (id, sku, item_name, category_id, base_unit_id, description, created_at, updated_at)
		VALUES (:id, :sku, :item_name, :category_id, :base_unit_id, :description, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, item)
	return err
}

func GetItem(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Item, error) {
	var item model.Item
	err := sqlx.GetContext(ctx, db, &item, "SELECT * FROM items WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &item, err
}

func ListItems(ctx context.Context, db sqlx.ExtContext, limit, offset int) ([]model.Item, error) {
	var items []model.Item
	err := sqlx.SelectContext(ctx, db, &items, "SELECT * FROM items ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return items, err
}

func UpdateItem(ctx context.Context, db sqlx.ExtContext, item *model.Item) error {
	query := `UPDATE items SET sku = $1, item_name = $2, category_id = $3, base_unit_id = $4, description = $5, updated_at = $6 WHERE id = $7`
	_, err := db.ExecContext(ctx, query, item.SKU, item.ItemName, item.CategoryID, item.BaseUnitID, item.Description, item.UpdatedAt, item.ID)
	return err
}

func DeleteItem(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM items WHERE id = $1", id)
	return err
}

func CreateUnit(ctx context.Context, db sqlx.ExtContext, u *model.UnitOfMeasure) error {
	query := `INSERT INTO units_of_measure (id, name, symbol, created_at)
		VALUES (:id, :name, :symbol, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, u)
	return err
}

func ListUnits(ctx context.Context, db sqlx.ExtContext) ([]model.UnitOfMeasure, error) {
	var units []model.UnitOfMeasure
	err := sqlx.SelectContext(ctx, db, &units, "SELECT * FROM units_of_measure ORDER BY name")
	return units, err
}

func CreateConversion(ctx context.Context, db sqlx.ExtContext, c *model.UnitConversion) error {
	query := `INSERT INTO unit_conversions (id, from_unit_id, to_unit_id, factor)
		VALUES (:id, :from_unit_id, :to_unit_id, :factor)`
	_, err := sqlx.NamedExecContext(ctx, db, query, c)
	return err
}

func ListConversions(ctx context.Context, db sqlx.ExtContext) ([]model.UnitConversion, error) {
	var convs []model.UnitConversion
	err := sqlx.SelectContext(ctx, db, &convs, "SELECT * FROM unit_conversions")
	return convs, err
}
