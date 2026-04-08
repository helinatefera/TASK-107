package service

import (
	"context"
	"time"

	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/masking"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateOrg(ctx context.Context, db sqlx.ExtContext, req *dto.CreateOrgRequest) (*model.Organization, error) {
	now := time.Now().UTC()
	org := &model.Organization{
		ID:        uuid.New(),
		ParentID:  req.ParentID,
		OrgCode:   req.OrgCode,
		Name:      req.Name,
		TaxID:     req.TaxID,
		Address:   req.Address,
		Timezone:  req.Timezone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateOrg(ctx, db, org); err != nil {
		return nil, err
	}
	return org, nil
}

func GetOrg(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Organization, error) {
	return repo.GetOrgByID(ctx, db, id)
}

func ListOrgs(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Organization, error) {
	if orgID != nil {
		return repo.ListOrgsByOrgID(ctx, db, *orgID, limit, offset)
	}
	return repo.ListOrgs(ctx, db, limit, offset)
}

func UpdateOrg(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateOrgRequest) (*model.Organization, error) {
	org, err := repo.GetOrgByID(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.TaxID != nil {
		org.TaxID = req.TaxID
	}
	if req.Address != nil {
		org.Address = req.Address
	}
	if req.Timezone != nil {
		org.Timezone = *req.Timezone
	}

	org.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateOrg(ctx, db, org); err != nil {
		return nil, err
	}
	return org, nil
}

func DeleteOrg(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteOrg(ctx, db, id)
}

func MaskOrg(org *model.Organization, role string) dto.OrgResponse {
	taxID := ""
	if org.TaxID != nil {
		taxID = masking.MaskTaxID(*org.TaxID, role)
	}
	address := ""
	if org.Address != nil {
		address = masking.MaskAddress(*org.Address, role)
	}
	return dto.OrgResponse{
		ID:        org.ID,
		ParentID:  org.ParentID,
		OrgCode:   org.OrgCode,
		Name:      org.Name,
		TaxID:     taxID,
		Address:   address,
		Timezone:  org.Timezone,
		CreatedAt: org.CreatedAt.Format(time.RFC3339),
		UpdatedAt: org.UpdatedAt.Format(time.RFC3339),
	}
}
