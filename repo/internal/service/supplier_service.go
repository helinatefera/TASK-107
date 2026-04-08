package service

import (
	"context"
	"strings"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/masking"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateSupplier(ctx context.Context, db sqlx.ExtContext, req *dto.CreateSupplierRequest) (*model.Supplier, error) {
	normalized := strings.ToLower(strings.TrimSpace(req.Name))

	dup, err := repo.CheckSupplierDuplicate(ctx, db, &req.OrgID, normalized, req.TaxID, nil)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, apperror.ErrDuplicate
	}

	now := time.Now().UTC()
	s := &model.Supplier{
		ID:             uuid.New(),
		OrgID:          &req.OrgID,
		Name:           req.Name,
		NormalizedName: normalized,
		TaxID:          req.TaxID,
		ContactEmail:   req.ContactEmail,
		Address:        req.Address,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := repo.CreateSupplier(ctx, db, s); err != nil {
		return nil, err
	}
	return s, nil
}

func GetSupplier(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Supplier, error) {
	return repo.GetSupplier(ctx, db, id)
}

func ListSuppliers(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Supplier, error) {
	return repo.ListSuppliers(ctx, db, orgID, limit, offset)
}

func UpdateSupplier(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateSupplierRequest) (*model.Supplier, error) {
	s, err := repo.GetSupplier(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		s.Name = *req.Name
		s.NormalizedName = strings.ToLower(strings.TrimSpace(*req.Name))
	}
	if req.TaxID != nil {
		s.TaxID = req.TaxID
	}
	if req.ContactEmail != nil {
		s.ContactEmail = req.ContactEmail
	}
	if req.Address != nil {
		s.Address = req.Address
	}

	// Re-check duplicates on name/taxID changes (exclude self)
	if req.Name != nil || req.TaxID != nil {
		dup, err := repo.CheckSupplierDuplicate(ctx, db, s.OrgID, s.NormalizedName, s.TaxID, &id)
		if err != nil {
			return nil, err
		}
		if dup {
			return nil, apperror.ErrDuplicate
		}
	}

	s.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateSupplier(ctx, db, s); err != nil {
		return nil, err
	}
	return s, nil
}

func MaskSupplier(s *model.Supplier, role string) {
	if s.TaxID != nil {
		masked := masking.MaskTaxID(*s.TaxID, role)
		s.TaxID = &masked
	}
	if s.Address != nil {
		masked := masking.MaskAddress(*s.Address, role)
		s.Address = &masked
	}
}

func CreateCarrier(ctx context.Context, db sqlx.ExtContext, req *dto.CreateCarrierRequest) (*model.Carrier, error) {
	normalized := strings.ToLower(strings.TrimSpace(req.Name))

	dup, err := repo.CheckCarrierDuplicate(ctx, db, &req.OrgID, normalized, req.TaxID, nil)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, apperror.ErrDuplicate
	}

	now := time.Now().UTC()
	c := &model.Carrier{
		ID:             uuid.New(),
		OrgID:          &req.OrgID,
		Name:           req.Name,
		NormalizedName: normalized,
		TaxID:          req.TaxID,
		ContactEmail:   req.ContactEmail,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := repo.CreateCarrier(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func GetCarrier(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Carrier, error) {
	return repo.GetCarrier(ctx, db, id)
}

func ListCarriers(ctx context.Context, db sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.Carrier, error) {
	return repo.ListCarriers(ctx, db, orgID, limit, offset)
}

func UpdateCarrier(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateCarrierRequest) (*model.Carrier, error) {
	c, err := repo.GetCarrier(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		c.Name = *req.Name
		c.NormalizedName = strings.ToLower(strings.TrimSpace(*req.Name))
	}
	if req.TaxID != nil {
		c.TaxID = req.TaxID
	}
	if req.ContactEmail != nil {
		c.ContactEmail = req.ContactEmail
	}

	// Re-check duplicates on name/taxID changes (exclude self)
	if req.Name != nil || req.TaxID != nil {
		dup, err := repo.CheckCarrierDuplicate(ctx, db, c.OrgID, c.NormalizedName, c.TaxID, &id)
		if err != nil {
			return nil, err
		}
		if dup {
			return nil, apperror.ErrDuplicate
		}
	}

	c.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateCarrier(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func MaskCarrier(c *model.Carrier, role string) {
	if c.TaxID != nil {
		masked := masking.MaskTaxID(*c.TaxID, role)
		c.TaxID = &masked
	}
}
