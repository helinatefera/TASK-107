package middleware

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type MetricEntry struct {
	Method     string
	Path       string
	StatusCode int
	LatencyMs  int64
}

var metricsCh chan MetricEntry

func MetricsRecorder(db *sqlx.DB) echo.MiddlewareFunc {
	metricsCh = make(chan MetricEntry, 1000)
	go StartMetricsFlusher(db, metricsCh)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			entry := MetricEntry{
				Method:     c.Request().Method,
				Path:       c.Request().URL.Path,
				StatusCode: c.Response().Status,
				LatencyMs:  time.Since(start).Milliseconds(),
			}

			select {
			case metricsCh <- entry:
			default:
				// channel full, drop metric
			}

			return err
		}
	}
}

func StartMetricsFlusher(db *sqlx.DB, ch <-chan MetricEntry) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var buffer []MetricEntry

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				if len(buffer) > 0 {
					flushMetrics(db, buffer)
				}
				return
			}
			buffer = append(buffer, entry)
		case <-ticker.C:
			if len(buffer) == 0 {
				continue
			}
			flushMetrics(db, buffer)
			buffer = nil
		}
	}
}

func flushMetrics(db *sqlx.DB, entries []MetricEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("metrics flush: begin tx failed")
		return
	}

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO request_metrics (method, path, status_code, latency_ms, recorded_at) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Error().Err(err).Msg("metrics flush: prepare failed")
		_ = tx.Rollback()
		return
	}
	defer stmt.Close()

	now := time.Now()
	for _, e := range entries {
		_, err = stmt.ExecContext(ctx, e.Method, e.Path, e.StatusCode, e.LatencyMs, now)
		if err != nil {
			log.Error().Err(err).Msg("metrics flush: insert failed")
			_ = tx.Rollback()
			return
		}
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("metrics flush: commit failed")
	}
}
