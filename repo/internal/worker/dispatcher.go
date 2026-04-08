package worker

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func StartAll(ctx context.Context, db *sqlx.DB) {
	log.Info().Msg("starting background workers")
	go StartNotificationWorker(ctx, db)
	go StartCleanupWorker(ctx, db)
}
