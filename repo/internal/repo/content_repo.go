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

func CreateCarousel(ctx context.Context, db sqlx.ExtContext, c *model.CarouselSlot) error {
	query := `INSERT INTO carousel_slots (id, org_id, title, image_url, link_url, priority, target_role, start_time, end_time, active, created_at, updated_at)
		VALUES (:id, :org_id, :title, :image_url, :link_url, :priority, :target_role, :start_time, :end_time, :active, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, c)
	return err
}

func GetCarousel(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.CarouselSlot, error) {
	var c model.CarouselSlot
	err := sqlx.GetContext(ctx, db, &c, "SELECT * FROM carousel_slots WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &c, err
}

func ListCarousels(ctx context.Context, db sqlx.ExtContext, role string, orgID *uuid.UUID) ([]model.CarouselSlot, error) {
	var slots []model.CarouselSlot
	if orgID != nil {
		query := `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id)
			SELECT cs.* FROM carousel_slots cs JOIN org_tree ot ON cs.org_id = ot.id
			WHERE cs.active = TRUE
			AND (cs.target_role IS NULL OR cs.target_role = $2)
			AND cs.start_time <= NOW() AND cs.end_time >= NOW()
			ORDER BY cs.priority`
		err := sqlx.SelectContext(ctx, db, &slots, query, *orgID, role)
		return slots, err
	}
	query := `SELECT * FROM carousel_slots
		WHERE active = TRUE
		AND (target_role IS NULL OR target_role = $1)
		AND start_time <= NOW() AND end_time >= NOW()
		ORDER BY priority`
	err := sqlx.SelectContext(ctx, db, &slots, query, role)
	return slots, err
}

func UpdateCarousel(ctx context.Context, db sqlx.ExtContext, c *model.CarouselSlot) error {
	query := `UPDATE carousel_slots SET org_id = $1, title = $2, image_url = $3, link_url = $4, priority = $5, target_role = $6, start_time = $7, end_time = $8, active = $9, updated_at = $10 WHERE id = $11`
	_, err := db.ExecContext(ctx, query, c.OrgID, c.Title, c.ImageURL, c.LinkURL, c.Priority, c.TargetRole, c.StartTime, c.EndTime, c.Active, c.UpdatedAt, c.ID)
	return err
}

func DeleteCarousel(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM carousel_slots WHERE id = $1", id)
	return err
}

func CreateCampaign(ctx context.Context, db sqlx.ExtContext, c *model.CampaignPlacement) error {
	query := `INSERT INTO campaign_placements (id, org_id, name, content, priority, target_role, start_time, end_time, active, created_at, updated_at)
		VALUES (:id, :org_id, :name, :content, :priority, :target_role, :start_time, :end_time, :active, :created_at, :updated_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, c)
	return err
}

func GetCampaign(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.CampaignPlacement, error) {
	var c model.CampaignPlacement
	err := sqlx.GetContext(ctx, db, &c, "SELECT * FROM campaign_placements WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &c, err
}

func ListCampaigns(ctx context.Context, db sqlx.ExtContext, role string, orgID *uuid.UUID) ([]model.CampaignPlacement, error) {
	var campaigns []model.CampaignPlacement
	if orgID != nil {
		query := `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id)
			SELECT cp.* FROM campaign_placements cp JOIN org_tree ot ON cp.org_id = ot.id
			WHERE cp.active = TRUE
			AND (cp.target_role IS NULL OR cp.target_role = $2)
			AND cp.start_time <= NOW() AND cp.end_time >= NOW()
			ORDER BY cp.priority`
		err := sqlx.SelectContext(ctx, db, &campaigns, query, *orgID, role)
		return campaigns, err
	}
	query := `SELECT * FROM campaign_placements
		WHERE active = TRUE
		AND (target_role IS NULL OR target_role = $1)
		AND start_time <= NOW() AND end_time >= NOW()
		ORDER BY priority`
	err := sqlx.SelectContext(ctx, db, &campaigns, query, role)
	return campaigns, err
}

func UpdateCampaign(ctx context.Context, db sqlx.ExtContext, c *model.CampaignPlacement) error {
	query := `UPDATE campaign_placements SET org_id = $1, name = $2, content = $3, priority = $4, target_role = $5, start_time = $6, end_time = $7, active = $8, updated_at = $9 WHERE id = $10`
	_, err := db.ExecContext(ctx, query, c.OrgID, c.Name, c.Content, c.Priority, c.TargetRole, c.StartTime, c.EndTime, c.Active, c.UpdatedAt, c.ID)
	return err
}

func DeleteCampaign(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM campaign_placements WHERE id = $1", id)
	return err
}

func CreateRanking(ctx context.Context, db sqlx.ExtContext, r *model.HotRanking) error {
	query := `INSERT INTO hot_rankings (id, org_id, entity_type, entity_id, score, priority, target_role, start_time, end_time, active, created_at)
		VALUES (:id, :org_id, :entity_type, :entity_id, :score, :priority, :target_role, :start_time, :end_time, :active, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, r)
	return err
}

func GetRanking(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.HotRanking, error) {
	var r model.HotRanking
	err := sqlx.GetContext(ctx, db, &r, "SELECT * FROM hot_rankings WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &r, err
}

func ListRankings(ctx context.Context, db sqlx.ExtContext, role string, orgID *uuid.UUID) ([]model.HotRanking, error) {
	var rankings []model.HotRanking
	if orgID != nil {
		query := `WITH RECURSIVE org_tree AS (SELECT id FROM organizations WHERE id = $1 UNION ALL SELECT o.id FROM organizations o JOIN org_tree t ON o.parent_id = t.id)
			SELECT hr.* FROM hot_rankings hr JOIN org_tree ot ON hr.org_id = ot.id
			WHERE hr.active = TRUE
			AND (hr.target_role IS NULL OR hr.target_role = $2)
			AND hr.start_time <= NOW() AND hr.end_time >= NOW()
			ORDER BY hr.priority, hr.score DESC`
		err := sqlx.SelectContext(ctx, db, &rankings, query, *orgID, role)
		return rankings, err
	}
	query := `SELECT * FROM hot_rankings
		WHERE active = TRUE
		AND (target_role IS NULL OR target_role = $1)
		AND start_time <= NOW() AND end_time >= NOW()
		ORDER BY priority, score DESC`
	err := sqlx.SelectContext(ctx, db, &rankings, query, role)
	return rankings, err
}

func UpdateRanking(ctx context.Context, db sqlx.ExtContext, r *model.HotRanking) error {
	query := `UPDATE hot_rankings SET org_id = $1, entity_type = $2, entity_id = $3, score = $4, priority = $5, target_role = $6, start_time = $7, end_time = $8, active = $9 WHERE id = $10`
	_, err := db.ExecContext(ctx, query, r.OrgID, r.EntityType, r.EntityID, r.Score, r.Priority, r.TargetRole, r.StartTime, r.EndTime, r.Active, r.ID)
	return err
}

func DeleteRanking(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "DELETE FROM hot_rankings WHERE id = $1", id)
	return err
}
