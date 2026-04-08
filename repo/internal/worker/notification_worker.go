package worker

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func StartNotificationWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processBatch(ctx, db)
		}
	}
}

func processBatch(ctx context.Context, db *sqlx.DB) {
	jobs, err := repo.ClaimPendingJobs(ctx, db, 50)
	if err != nil {
		log.Error().Err(err).Msg("failed to claim notification jobs")
		return
	}
	for _, job := range jobs {
		func(j model.NotificationJob) {
			jobCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			processJob(jobCtx, db, j)
		}(job)
	}
}

func processJob(ctx context.Context, db *sqlx.DB, job model.NotificationJob) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Int64("job_id", job.ID).Interface("panic", r).Msg("panic in processJob")
			repo.UpdateJobStatus(ctx, db, job.ID, "failed", "panic")
		}
	}()

	// 1. Check opt-out
	optedOut, err := repo.IsOptedOut(ctx, db, job.UserID, job.TemplateID)
	if err != nil {
		log.Error().Err(err).Int64("job_id", job.ID).Msg("failed to check opt-out")
		repo.UpdateJobStatus(ctx, db, job.ID, "failed", "opt_out_check_error")
		return
	}
	if optedOut {
		if err := repo.UpdateJobStatus(ctx, db, job.ID, "suppressed", "opted_out"); err != nil {
			log.Error().Err(err).Int64("job_id", job.ID).Msg("failed to update suppressed status")
		}
		return
	}

	// 2. Check quiet hours (9 PM - 7 AM local time)
	userTZ := getUserTimezone(ctx, db, job.UserID)
	loc, err := time.LoadLocation(userTZ)
	if err != nil {
		loc = time.UTC
	}
	localNow := time.Now().In(loc)
	if IsQuietHours(localNow.Hour()) {
		next7AM := ComputeNext7AM(localNow, loc)
		repo.RescheduleJob(ctx, db, job.ID, next7AM)
		return
	}

	// 3. Check rate limit: max 2 per user per template per day (user-local day)
	// Use time.Date to get midnight in the user's local timezone (Truncate uses UTC)
	y, m, d := localNow.Date()
	localDay := time.Date(y, m, d, 0, 0, 0, 0, loc)
	nextDay := localDay.AddDate(0, 0, 1)
	count, err := repo.CountDeliveredInWindow(ctx, db, job.UserID, job.TemplateID, localDay.UTC(), nextDay.UTC())
	if err != nil {
		log.Error().Err(err).Int64("job_id", job.ID).Msg("failed to count delivered today")
		repo.UpdateJobStatus(ctx, db, job.ID, "failed", "rate_limit_check_error")
		return
	}
	if count >= 2 {
		repo.UpdateJobStatus(ctx, db, job.ID, "suppressed", "rate_limit")
		return
	}

	// 4. Render and deliver to inbox
	tmpl, err := repo.GetNotificationTemplate(ctx, db, job.TemplateID)
	if err != nil {
		log.Error().Err(err).Int64("job_id", job.ID).Msg("failed to get template")
		repo.UpdateJobStatus(ctx, db, job.ID, "failed", "template_not_found")
		return
	}

	var params []byte
	if job.Params != nil {
		params = *job.Params
	}
	title := RenderTemplate(tmpl.TitleTmpl, params)
	body := RenderTemplate(tmpl.BodyTmpl, params)

	msg := &model.Message{
		ID:         uuid.New(),
		UserID:     job.UserID,
		TemplateID: job.TemplateID,
		Title:      title,
		Body:       body,
		CreatedAt:  time.Now().UTC(),
	}
	if err := repo.CreateMessage(ctx, db, msg); err != nil {
		log.Error().Err(err).Int64("job_id", job.ID).Msg("failed to create message")
		repo.UpdateJobStatus(ctx, db, job.ID, "failed", "delivery_failed")
		return
	}

	repo.UpdateJobStatus(ctx, db, job.ID, "delivered", "")

	// 5. Increment delivered stat
	repo.IncrementDeliveryStat(ctx, db, job.TemplateID, "delivered")
}

func getUserTimezone(ctx context.Context, db sqlx.ExtContext, userID interface{}) string {
	var tz string
	err := sqlx.GetContext(ctx, db, &tz,
		`SELECT COALESCE(o.timezone, 'UTC') FROM users u
		 LEFT JOIN organizations o ON o.id = u.org_id
		 WHERE u.id = $1`, userID)
	if err != nil {
		return "UTC"
	}
	return tz
}

// IsQuietHours returns true if the given hour falls within quiet hours (9 PM - 7 AM).
func IsQuietHours(hour int) bool {
	return hour >= 21 || hour < 7
}

func ComputeNext7AM(now time.Time, loc *time.Location) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, loc)
	if now.Hour() >= 7 {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func RenderTemplate(tmpl string, params []byte) string {
	if len(params) == 0 {
		return tmpl
	}
	result := tmpl
	// Simple key-value replacement from JSON params
	var m map[string]string
	if err := json.Unmarshal(params, &m); err != nil {
		return tmpl
	}
	for k, v := range m {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}
