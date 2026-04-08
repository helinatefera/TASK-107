package worker

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func StartCleanupWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanExpiredSessions(ctx, db)
			cleanExpiredRecoveryTokens(ctx, db)
			purgeOldMetrics(ctx, db)
		}
	}
}

func cleanExpiredSessions(ctx context.Context, db *sqlx.DB) {
	result, err := db.ExecContext(ctx,
		"DELETE FROM sessions WHERE abs_expires_at < now()")
	if err != nil {
		log.Error().Err(err).Msg("failed to clean expired sessions")
		return
	}
	if n, _ := result.RowsAffected(); n > 0 {
		log.Info().Int64("count", n).Msg("cleaned expired sessions")
	}
}

// purgeOldMetrics enforces a 30-day retention policy for request metrics.
func purgeOldMetrics(ctx context.Context, db *sqlx.DB) {
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	result, err := db.ExecContext(ctx,
		"DELETE FROM request_metrics WHERE recorded_at < $1", cutoff)
	if err != nil {
		log.Error().Err(err).Msg("failed to purge old metrics")
		return
	}
	if n, _ := result.RowsAffected(); n > 0 {
		log.Info().Int64("count", n).Msg("purged old request metrics (>30 days)")
	}
}

func cleanExpiredRecoveryTokens(ctx context.Context, db *sqlx.DB) {
	result, err := db.ExecContext(ctx,
		"DELETE FROM recovery_tokens WHERE expires_at < now() OR used = TRUE")
	if err != nil {
		log.Error().Err(err).Msg("failed to clean expired recovery tokens")
		return
	}
	if n, _ := result.RowsAffected(); n > 0 {
		log.Info().Int64("count", n).Msg("cleaned expired recovery tokens")
	}
}
