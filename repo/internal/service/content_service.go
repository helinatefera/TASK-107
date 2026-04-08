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
)

// --- Carousel ---

func CreateCarousel(ctx context.Context, db sqlx.ExtContext, req *dto.CreateCarouselRequest) (*model.CarouselSlot, error) {
	if !req.StartTime.Before(req.EndTime) {
		return nil, apperror.ErrInvalidTimeRange
	}

	now := time.Now().UTC()
	c := &model.CarouselSlot{
		ID:         uuid.New(),
		OrgID:      req.OrgID,
		Title:      req.Title,
		ImageURL:   req.ImageURL,
		LinkURL:    req.LinkURL,
		Priority:   req.Priority,
		TargetRole: req.TargetRole,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Active:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.CreateCarousel(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func GetCarousel(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.CarouselSlot, error) {
	return repo.GetCarousel(ctx, db, id)
}

func ListCarousels(ctx context.Context, db sqlx.ExtContext, role string, orgID *uuid.UUID) ([]model.CarouselSlot, error) {
	return repo.ListCarousels(ctx, db, role, orgID)
}

func UpdateCarousel(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateCarouselRequest) (*model.CarouselSlot, error) {
	c, err := repo.GetCarousel(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		c.Title = *req.Title
	}
	if req.ImageURL != nil {
		c.ImageURL = req.ImageURL
	}
	if req.LinkURL != nil {
		c.LinkURL = req.LinkURL
	}
	if req.Priority != nil {
		c.Priority = *req.Priority
	}
	if req.TargetRole != nil {
		c.TargetRole = req.TargetRole
	}
	if req.StartTime != nil {
		c.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		c.EndTime = *req.EndTime
	}
	if req.Active != nil {
		c.Active = *req.Active
	}

	if !c.StartTime.Before(c.EndTime) {
		return nil, apperror.ErrInvalidTimeRange
	}

	c.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateCarousel(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func DeleteCarousel(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteCarousel(ctx, db, id)
}

// --- Campaign ---

func CreateCampaign(ctx context.Context, db sqlx.ExtContext, req *dto.CreateCampaignRequest) (*model.CampaignPlacement, error) {
	if !req.StartTime.Before(req.EndTime) {
		return nil, apperror.ErrInvalidTimeRange
	}

	now := time.Now().UTC()
	c := &model.CampaignPlacement{
		ID:         uuid.New(),
		OrgID:      req.OrgID,
		Name:       req.Name,
		Content:    req.Content,
		Priority:   req.Priority,
		TargetRole: req.TargetRole,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Active:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.CreateCampaign(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func GetCampaign(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.CampaignPlacement, error) {
	return repo.GetCampaign(ctx, db, id)
}

func ListCampaigns(ctx context.Context, db sqlx.ExtContext, role string, orgID *uuid.UUID) ([]model.CampaignPlacement, error) {
	return repo.ListCampaigns(ctx, db, role, orgID)
}

func UpdateCampaign(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateCampaignRequest) (*model.CampaignPlacement, error) {
	c, err := repo.GetCampaign(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Content != nil {
		c.Content = *req.Content
	}
	if req.Priority != nil {
		c.Priority = *req.Priority
	}
	if req.TargetRole != nil {
		c.TargetRole = req.TargetRole
	}
	if req.StartTime != nil {
		c.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		c.EndTime = *req.EndTime
	}
	if req.Active != nil {
		c.Active = *req.Active
	}

	if !c.StartTime.Before(c.EndTime) {
		return nil, apperror.ErrInvalidTimeRange
	}

	c.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateCampaign(ctx, db, c); err != nil {
		return nil, err
	}
	return c, nil
}

func DeleteCampaign(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteCampaign(ctx, db, id)
}

// --- Rankings ---

func CreateRanking(ctx context.Context, db sqlx.ExtContext, req *dto.CreateRankingRequest) (*model.HotRanking, error) {
	if !req.StartTime.Before(req.EndTime) {
		return nil, apperror.ErrInvalidTimeRange
	}

	r := &model.HotRanking{
		ID:         uuid.New(),
		OrgID:      req.OrgID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Score:      req.Score,
		Priority:   req.Priority,
		TargetRole: req.TargetRole,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Active:     true,
		CreatedAt:  time.Now().UTC(),
	}
	if err := repo.CreateRanking(ctx, db, r); err != nil {
		return nil, err
	}
	return r, nil
}

func GetRanking(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.HotRanking, error) {
	return repo.GetRanking(ctx, db, id)
}

func ListRankings(ctx context.Context, db sqlx.ExtContext, role string, orgID *uuid.UUID) ([]model.HotRanking, error) {
	return repo.ListRankings(ctx, db, role, orgID)
}

func UpdateRanking(ctx context.Context, db sqlx.ExtContext, id uuid.UUID, req *dto.UpdateRankingRequest) (*model.HotRanking, error) {
	r, err := repo.GetRanking(ctx, db, id)
	if err != nil {
		return nil, err
	}

	if req.Score != nil {
		r.Score = *req.Score
	}
	if req.Priority != nil {
		r.Priority = *req.Priority
	}
	if req.TargetRole != nil {
		r.TargetRole = req.TargetRole
	}
	if req.StartTime != nil {
		r.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		r.EndTime = *req.EndTime
	}
	if req.Active != nil {
		r.Active = *req.Active
	}

	if !r.StartTime.Before(r.EndTime) {
		return nil, apperror.ErrInvalidTimeRange
	}

	if err := repo.UpdateRanking(ctx, db, r); err != nil {
		return nil, err
	}
	return r, nil
}

func DeleteRanking(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteRanking(ctx, db, id)
}
