package apperror

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (e *AppError) Error() string { return e.Message }

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

var (
	ErrNotFound           = New(http.StatusNotFound, "resource not found")
	ErrUnauthorized       = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden          = New(http.StatusForbidden, "forbidden")
	ErrBadRequest         = New(http.StatusBadRequest, "bad request")
	ErrConflict           = New(http.StatusConflict, "resource already exists")
	ErrTooManyRequests    = New(http.StatusTooManyRequests, "account locked, try again later")
	ErrPasswordTooShort   = New(http.StatusBadRequest, "password must be at least 12 characters")
	ErrPasswordComplexity = New(http.StatusBadRequest, "password must include 3 of 4 character classes: lowercase, uppercase, digits, special")
	ErrInvalidToken       = New(http.StatusBadRequest, "invalid or expired token")
	ErrAccountLocked      = New(http.StatusTooManyRequests, "account temporarily locked due to too many failed attempts")
	ErrSessionExpired     = New(http.StatusUnauthorized, "session expired")
	ErrDeviceMismatch     = New(http.StatusUnauthorized, "device mismatch")
	ErrDuplicate          = New(http.StatusConflict, "duplicate entry detected")
	ErrTOUOverlap         = New(http.StatusBadRequest, "time-of-use windows must not overlap for the same day type")
	ErrInvalidTimeRange   = New(http.StatusBadRequest, "start time must be before end time")
	ErrInternal           = New(http.StatusInternalServerError, "internal server error")
)

// HTTPErrorHandler returns consistent JSON error responses.
// Never exposes stack traces, internal details, or framework internals.
func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if ae, ok := err.(*AppError); ok {
		c.JSON(ae.Code, ae)
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		msg := "request error"
		if m, ok := he.Message.(string); ok {
			msg = m
		}
		c.JSON(he.Code, &AppError{Code: he.Code, Message: msg})
		return
	}

	// Catch PostgreSQL unique constraint violations → 409 Conflict.
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		c.JSON(http.StatusConflict, ErrDuplicate)
		return
	}

	// Log the real error internally but never expose it
	log.Error().Err(err).Str("path", c.Request().URL.Path).Msg("unhandled error")
	c.JSON(http.StatusInternalServerError, ErrInternal)
}
