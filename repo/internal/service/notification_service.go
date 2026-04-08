package service

import (
	"context"
	"time"

	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateNotificationTemplate(ctx context.Context, db sqlx.ExtContext, req *dto.CreateNotificationTemplateRequest) (*model.NotificationTemplate, error) {
	defaultEnabled := true
	if req.DefaultEnabled != nil {
		defaultEnabled = *req.DefaultEnabled
	}

	t := &model.NotificationTemplate{
		ID:             uuid.New(),
		Code:           req.Code,
		TitleTmpl:      req.TitleTmpl,
		BodyTmpl:       req.BodyTmpl,
		DefaultEnabled: defaultEnabled,
		CreatedAt:      time.Now().UTC(),
	}
	if err := repo.CreateNotificationTemplate(ctx, db, t); err != nil {
		return nil, err
	}
	return t, nil
}

func ListNotificationTemplates(ctx context.Context, db sqlx.ExtContext) ([]model.NotificationTemplate, error) {
	return repo.ListNotificationTemplates(ctx, db)
}

func GetSubscriptions(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID) ([]model.NotificationSubscription, error) {
	return repo.ListSubscriptions(ctx, db, userID)
}

func UpdateSubscription(ctx context.Context, db sqlx.ExtContext, userID, templateID uuid.UUID, optedOut bool) error {
	sub := &model.NotificationSubscription{
		ID:         uuid.New(),
		UserID:     userID,
		TemplateID: templateID,
		OptedOut:   optedOut,
		UpdatedAt:  time.Now().UTC(),
	}
	return repo.UpsertSubscription(ctx, db, sub)
}

func SendNotification(ctx context.Context, db sqlx.ExtContext, req *dto.SendNotificationRequest) error {
	now := time.Now().UTC()
	job := &model.NotificationJob{
		UserID:      req.UserID,
		TemplateID:  req.TemplateID,
		Params:      &req.Params,
		Status:      "pending",
		ScheduledAt: now,
		CreatedAt:   now,
	}
	if err := repo.CreateJob(ctx, db, job); err != nil {
		return err
	}

	if err := repo.IncrementDeliveryStat(ctx, db, req.TemplateID, "generated"); err != nil {
		return err
	}

	return nil
}

func GetInbox(ctx context.Context, db sqlx.ExtContext, userID uuid.UUID, limit, offset int) ([]model.Message, error) {
	return repo.ListMessages(ctx, db, userID, limit, offset)
}

func MarkRead(ctx context.Context, db sqlx.ExtContext, messageID uuid.UUID) error {
	msg, err := repo.GetMessage(ctx, db, messageID)
	if err != nil {
		return err
	}

	// Only increment stat if not already read (prevent double-counting)
	if msg.Read {
		return nil
	}

	if err := repo.MarkMessageRead(ctx, db, messageID); err != nil {
		return err
	}

	return repo.IncrementDeliveryStat(ctx, db, msg.TemplateID, "opened")
}

func MarkDismissed(ctx context.Context, db sqlx.ExtContext, messageID uuid.UUID) error {
	msg, err := repo.GetMessage(ctx, db, messageID)
	if err != nil {
		return err
	}

	// Only increment stat if not already dismissed (prevent double-counting)
	if msg.Dismissed {
		return nil
	}

	if err := repo.MarkMessageDismissed(ctx, db, messageID); err != nil {
		return err
	}

	return repo.IncrementDeliveryStat(ctx, db, msg.TemplateID, "dismissed")
}

func GetDeliveryStats(ctx context.Context, db sqlx.ExtContext, templateID *uuid.UUID) ([]model.DeliveryStats, error) {
	return repo.ListDeliveryStats(ctx, db, templateID)
}
