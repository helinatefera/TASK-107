package unit_tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chargeops/api/internal/apperror"
	"github.com/labstack/echo/v4"
)

func TestAppErrorHasCorrectCodeAndMessage(t *testing.T) {
	err := apperror.New(http.StatusBadRequest, "something went wrong")
	if err.Code != http.StatusBadRequest {
		t.Fatalf("expected code %d, got %d", http.StatusBadRequest, err.Code)
	}
	if err.Message != "something went wrong" {
		t.Fatalf("expected message %q, got %q", "something went wrong", err.Message)
	}
}

func TestNewCreatesAppErrorWithCorrectFields(t *testing.T) {
	tests := []struct {
		code int
		msg  string
	}{
		{http.StatusNotFound, "not found"},
		{http.StatusInternalServerError, "server error"},
		{http.StatusForbidden, "forbidden"},
	}

	for _, tc := range tests {
		err := apperror.New(tc.code, tc.msg)
		if err.Code != tc.code {
			t.Errorf("New(%d, %q).Code = %d", tc.code, tc.msg, err.Code)
		}
		if err.Message != tc.msg {
			t.Errorf("New(%d, %q).Message = %q", tc.code, tc.msg, err.Message)
		}
	}
}

func TestErrorReturnsMessageString(t *testing.T) {
	err := apperror.New(http.StatusBadRequest, "bad request")
	if err.Error() != "bad request" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "bad request")
	}
}

func TestAppErrorImplementsErrorInterface(t *testing.T) {
	var err error = apperror.New(http.StatusBadRequest, "test")
	if err.Error() != "test" {
		t.Fatalf("expected error interface to return message")
	}
}

func setupEchoContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestHTTPErrorHandlerWithAppError(t *testing.T) {
	c, rec := setupEchoContext(http.MethodGet, "/test")

	appErr := apperror.New(http.StatusBadRequest, "invalid input")
	apperror.HTTPErrorHandler(appErr, c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if int(body["code"].(float64)) != http.StatusBadRequest {
		t.Fatalf("expected JSON code %d, got %v", http.StatusBadRequest, body["code"])
	}
	if body["msg"] != "invalid input" {
		t.Fatalf("expected JSON msg %q, got %v", "invalid input", body["msg"])
	}
}

func TestHTTPErrorHandlerWithEchoHTTPError(t *testing.T) {
	c, rec := setupEchoContext(http.MethodGet, "/test")

	echoErr := echo.NewHTTPError(http.StatusNotFound, "page not found")
	apperror.HTTPErrorHandler(echoErr, c)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if int(body["code"].(float64)) != http.StatusNotFound {
		t.Fatalf("expected JSON code %d, got %v", http.StatusNotFound, body["code"])
	}
	if body["msg"] != "page not found" {
		t.Fatalf("expected JSON msg %q, got %v", "page not found", body["msg"])
	}
}

func TestHTTPErrorHandlerWithGenericError(t *testing.T) {
	c, rec := setupEchoContext(http.MethodGet, "/test")

	genericErr := errors.New("database connection failed: password=secret123")
	apperror.HTTPErrorHandler(genericErr, c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if int(body["code"].(float64)) != http.StatusInternalServerError {
		t.Fatalf("expected JSON code 500, got %v", body["code"])
	}
	if body["msg"] != "internal server error" {
		t.Fatalf("expected generic message %q, got %v", "internal server error", body["msg"])
	}

	// Verify the actual error message is NOT exposed
	responseBody := rec.Body.String()
	if strings.Contains(responseBody, "database connection failed") {
		t.Fatal("response body should not contain internal error details")
	}
	if strings.Contains(responseBody, "secret123") {
		t.Fatal("response body should not contain sensitive data")
	}
}

func TestHTTPErrorHandlerNoStackTrace(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"AppError", apperror.New(http.StatusBadRequest, "bad")},
		{"echo.HTTPError", echo.NewHTTPError(http.StatusNotFound, "not found")},
		{"generic error", errors.New("something broke internally")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := setupEchoContext(http.MethodGet, "/test")
			apperror.HTTPErrorHandler(tc.err, c)

			body := rec.Body.String()
			if strings.Contains(body, "goroutine") {
				t.Fatal("response contains stack trace (goroutine)")
			}
			if strings.Contains(body, ".go:") {
				t.Fatal("response contains file reference (.go:)")
			}
			if strings.Contains(body, "runtime.") {
				t.Fatal("response contains runtime reference")
			}
		})
	}
}
