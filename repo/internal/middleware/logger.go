package middleware

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func Logger() echo.MiddlewareFunc {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			latency := time.Since(start)

			reqID, _ := c.Get("request_id").(string)

			logger.Info().
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Int("status", c.Response().Status).
				Int64("latency_ms", latency.Milliseconds()).
				Str("request_id", reqID).
				Str("remote_ip", c.RealIP()).
				Msg("request")

			return err
		}
	}
}
