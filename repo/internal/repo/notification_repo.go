package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateNotificationTemplate(ctx context.Context, db sqlx.ExtContext, t *model.NotificationTemplate) error {
	query := `INSERT INTO notification_templates (id, code, title_tmpl, body_tmpl, default_enabled, created_at)
		VALUES (:id, :code, :title_tmpl, :body_tmpl, :default_enabled, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, t)
	return err
}

func GetNotificationTemplate(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.NotificationTemplate, error) {
	var t model.NotificationTemplate
	err := sqlx.GetContext(ctx, db, &t, "SELECT * FROM notification_templates WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &t, err
}

func GetNotificationTemplateByCode(ctx context.Context, db sqlx.ExtContext, code string) (*model.NotificationTemplate, error) {
	var t model.NotificationTemplate
	err := sqlx.GetContext(ctx, db, &t, "SELECT * FROM notification_templates WHERE code = $1", code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &t, err
}

func ListNotificationTemplates(ctx context.Context, db sqlx.ExtContext) ([]model.NotificationTemplate, error) {
	var templates []model.NotificationTemplate
	err := sqlx.SelectContext(ctx, db, &templates, "SELECT * FROM notification_templates ORDER BY code")
	return templates, err
}

func GetSubscription(ctx context.Context, db sqlx.ExtContext, userID, templateID uuid.UUID) (*model.NotificationSubscription, error) {
	var s model.NotificationSubscription
	err := sqlx.GetContext(ctx, db, &s, "SELECT * FROM notification_subscriptions WHERE user_id = $1 AND template_id = $2", userID, templateID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &s, err
}

func UpsertSubscription(ctx context.Context, db sqlx.ExtContext, sub *model.NotificationSubscription) error {
	query := `INSERT INTO notification_subscriptions (id, user_id, template_id, opted_out, updated_at)
		VALUES (:id, :user_id, :template_id, :opted_out, :updated_at)
		ON CONFLICT (user_id, template_id) DO UPDATE SET opted_out = EXCLUDED.opted_out, updated_at = EXCLUDED.updated_at`
	_, err := sqlx.NamedExecContext(ctx, db, query, sub)
	return err
}

func ListSubscriptions(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID) ([]model.NotificationSubscription, error) {
	var subs []model.NotificationSubscription
	err := sqlx.SelectContext(ctx, db, &subs, "SELECT * FROM notification_subscriptions WHERE user_id = $1", userID)
	return subs, err
}

func IsOptedOut(ctx context.Context, db sqlx.ExtContext, userID, templateID uuid.UUID) (bool, error) {
	var optedOut bool
	err := sqlx.GetContext(ctx, db, &optedOut, "SELECT opted_out FROM notification_subscriptions WHERE user_id = $1 AND template_id = $2", userID, templateID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return optedOut, err
}

func CreateJob(ctx context.Context, db sqlx.ExtContext, job *model.NotificationJob) error {
	query := `INSERT INTO notification_jobs (user_id, template_id, params, status, scheduled_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	return sqlx.GetContext(ctx, db, &job.ID, query,
		job.UserID, job.TemplateID, job.Params, job.Status, job.ScheduledAt, job.CreatedAt)
}

func ClaimPendingJobs(ctx context.Context, db sqlx.ExtContext, limit int) ([]model.NotificationJob, error) {
	var jobs []model.NotificationJob
	query := `UPDATE notification_jobs SET status = 'processing', processed_at = NOW()
		WHERE id IN (
			SELECT id FROM notification_jobs
			WHERE status = 'pending' AND scheduled_at <= NOW()
			ORDER BY scheduled_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING *`
	err := sqlx.SelectContext(ctx, db, &jobs, query, limit)
	return jobs, err
}

func UpdateJobStatus(ctx context.Context, db sqlx.ExtContext, jobID int64, status string, reason string) error {
	_, err := db.ExecContext(ctx, "UPDATE notification_jobs SET status = $1, suppress_reason = $2, processed_at = NOW() WHERE id = $3", status, reason, jobID)
	return err
}

func RescheduleJob(ctx context.Context, db sqlx.ExtContext, jobID int64, scheduledAt time.Time) error {
	_, err := db.ExecContext(ctx, "UPDATE notification_jobs SET status = 'pending', scheduled_at = $1, processed_at = NULL WHERE id = $2", scheduledAt, jobID)
	return err
}


func CountDeliveredInWindow(ctx context.Context, db sqlx.ExtContext, userID, templateID uuid.UUID, from, to time.Time) (int, error) {
	var count int
	err := sqlx.GetContext(ctx, db, &count,
		"SELECT COUNT(*) FROM notification_jobs WHERE user_id = $1 AND template_id = $2 AND status = 'delivered' AND processed_at >= $3 AND processed_at < $4",
		userID, templateID, from, to)
	return count, err
}

func CreateMessage(ctx context.Context, db sqlx.ExtContext, m *model.Message) error {
	query := `INSERT INTO messages (id, user_id, template_id, title, body, read, dismissed, created_at)
		VALUES (:id, :user_id, :template_id, :title, :body, :read, :dismissed, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, db, query, m)
	return err
}

func ListMessages(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, limit, offset int) ([]model.Message, error) {
	var msgs []model.Message
	err := sqlx.SelectContext(ctx, db, &msgs,
		"SELECT * FROM messages WHERE user_id = $1 AND dismissed = FALSE ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		userID, limit, offset)
	return msgs, err
}

func GetMessage(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) (*model.Message, error) {
	var m model.Message
	err := sqlx.GetContext(ctx, db, &m, "SELECT * FROM messages WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &m, err
}

func MarkMessageRead(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "UPDATE messages SET read = TRUE WHERE id = $1", id)
	return err
}

func MarkMessageDismissed(ctx context.Context, db sqlx.ExtContext, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, "UPDATE messages SET dismissed = TRUE WHERE id = $1", id)
	return err
}

func IncrementDeliveryStat(ctx context.Context, db sqlx.ExtContext, templateID uuid.UUID, field string) error {
	query := `INSERT INTO delivery_stats (template_id, date, generated, delivered, opened, dismissed)
		VALUES ($1, CURRENT_DATE, 0, 0, 0, 0)
		ON CONFLICT (template_id, date) DO NOTHING`
	_, err := db.ExecContext(ctx, query, templateID)
	if err != nil {
		return err
	}
	updateQuery := `UPDATE delivery_stats SET ` + field + ` = ` + field + ` + 1 WHERE template_id = $1 AND date = CURRENT_DATE`
	_, err = db.ExecContext(ctx, updateQuery, templateID)
	return err
}

func ListDeliveryStats(ctx context.Context, db sqlx.ExtContext, templateID *uuid.UUID) ([]model.DeliveryStats, error) {
	var stats []model.DeliveryStats
	if templateID != nil {
		err := sqlx.SelectContext(ctx, db, &stats, "SELECT * FROM delivery_stats WHERE template_id = $1 ORDER BY date DESC", *templateID)
		return stats, err
	}
	err := sqlx.SelectContext(ctx, db, &stats, "SELECT * FROM delivery_stats ORDER BY date DESC")
	return stats, err
}
