package service

import (
	"context"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

func CreateCategory(ctx context.Context, db sqlx.ExtContext, req *dto.CreateCategoryRequest) (*model.Category, error) {
	c := &model.Category{
		ID:        uuid.New(),
		Name:      req.Name,
		ParentID:  req.ParentID,
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.CreateCategory(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func ListCategories(ctx context.Context, db sqlx.ExtContext) ([]model.Category, error) {
	return repo.ListCategories(ctx, db)
}

func CreateItem(ctx context.Context, db sqlx.ExtContext, req *dto.CreateItemRequest) (*model.Item, error) {
	now := time.Now().UTC()
	item := &model.Item{
		ID:          uuid.New(),
		SKU:         req.SKU,
		ItemName:    req.ItemName,
		CategoryID:  req.CategoryID,
		BaseUnitID:  req.BaseUnitID,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateItem(ctx, db, item); err != nil {
		return nil, err
	}
	return item, nil
}

func GetItem(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Item, error) {
	return repo.GetItem(ctx, db, id)
}

func ListItems(ctx context.Context, db sqlx.ExtContext, limit, offset int) ([]model.Item, error) {
	return repo.ListItems(ctx, db, limit, offset)
}

func UpdateItem(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateItemRequest) (*model.Item, error) {
	item, err := repo.GetItem(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.ItemName != nil {
		item.ItemName = *req.ItemName
	}
	if req.CategoryID != nil {
		item.CategoryID = req.CategoryID
	}
	if req.Description != nil {
		item.Description = req.Description
	}

	item.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateItem(ctx, db, item); err != nil {
		return nil, err
	}
	return item, nil
}

func DeleteItem(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteItem(ctx, db, id)
}

func CreateUnit(ctx context.Context, db sqlx.ExtContext, req *dto.CreateUnitRequest) (*model.UnitOfMeasure, error) {
	u := &model.UnitOfMeasure{
		ID:        uuid.New(),
		Name:      req.Name,
		Symbol:    req.Symbol,
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.CreateUnit(ctx, db, u); err != nil {
		return nil, err
	}
	return u, nil
}

func ListUnits(ctx context.Context, db sqlx.ExtContext) ([]model.UnitOfMeasure, error) {
	return repo.ListUnits(ctx, db)
}

func CreateConversion(ctx context.Context, db sqlx.ExtContext, req *dto.CreateConversionRequest) (*model.UnitConversion, error) {
	if req.Factor.LessThanOrEqual(decimal.Zero) {
		return nil, apperror.ErrBadRequest
	}

	c := &model.UnitConversion{
		ID:         uuid.New(),
		FromUnitID: req.FromUnitID,
		ToUnitID:   req.ToUnitID,
		Factor:     req.Factor,
	}
	if err := repo.CreateConversion(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func ListConversions(ctx context.Context, db sqlx.ExtContext) ([]model.UnitConversion, error) {
	return repo.ListConversions(ctx, db)
}
