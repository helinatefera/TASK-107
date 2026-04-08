package unit_tests

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/handler"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// testValidator wraps go-playground/validator for Echo, matching the production setup.
type testValidator struct {
	v *validator.Validate
}

func (tv *testValidator) Validate(i interface{}) error {
	if err := tv.v.Struct(i); err != nil {
		return apperror.New(http.StatusBadRequest, "validation failed")
	}
	return nil
}

// fakeAuthMiddleware injects caller identity values into the echo.Context,
// replicating what the real Auth middleware does after token validation.
func fakeAuthMiddleware(userID uuid.UUID, role string, orgID *uuid.UUID) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_id", userID)
			c.Set("role", role)
			c.Set("org_id", orgID)
			return next(c)
		}
	}
}

// newTestEchoForUsers creates an Echo instance with GetUser and UpdateUser
// handlers registered behind the fake auth middleware.
func newTestEchoForUsers(callerID uuid.UUID, role string, orgID *uuid.UUID) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = apperror.HTTPErrorHandler
	// Recover middleware catches nil-DB panics on paths that pass auth checks.
	// Those paths are expected to panic (no real DB), but we only care that
	// the response is NOT 403 — Recover turns panics into 500s.
	e.Use(echomw.Recover())
	cfg := config.Load()

	g := e.Group("", fakeAuthMiddleware(callerID, role, orgID))
	g.GET("/users/:id", handler.GetUserHandler(nil, cfg))
	g.PUT("/users/:id", handler.UpdateUserHandler(nil, cfg))

	return e
}

// ---------------------------------------------------------------------------
// enforceSelfOrAdmin — tested via GetUserHandler
// ---------------------------------------------------------------------------

func TestSelfOrAdmin_AdminBypasses(t *testing.T) {
	// Admin should pass the auth check regardless of target user ID.
	// The handler will then try to call the DB, which is nil, so it will
	// panic or error — but we never reach that because Go will return a
	// runtime error. We expect a non-403 response. An admin whose request
	// passes the auth gate but fails at the DB layer gets a 500, not 403.
	adminID := uuid.New()
	targetID := uuid.New() // different from admin

	e := newTestEchoForUsers(adminID, "administrator", nil)
	req := httptest.NewRequest(http.MethodGet, "/users/"+targetID.String(), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("admin should bypass self-or-admin check, got 403")
	}
	// We expect something other than 403 (likely 500 because DB is nil).
	// The key assertion: the auth gate did NOT block the admin.
}

func TestSelfOrAdmin_MatchingSelfPasses(t *testing.T) {
	callerID := uuid.New()

	e := newTestEchoForUsers(callerID, "user", nil)
	req := httptest.NewRequest(http.MethodGet, "/users/"+callerID.String(), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("caller accessing own record should pass self check, got 403")
	}
}

func TestSelfOrAdmin_NonMatchingUserForbidden(t *testing.T) {
	callerID := uuid.New()
	otherID := uuid.New()

	e := newTestEchoForUsers(callerID, "user", nil)
	req := httptest.NewRequest(http.MethodGet, "/users/"+otherID.String(), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin accessing another user should get 403, got %d", rec.Code)
	}
}

func TestSelfOrAdmin_MerchantNonMatchingForbidden(t *testing.T) {
	callerID := uuid.New()
	otherID := uuid.New()
	orgID := uuid.New()

	e := newTestEchoForUsers(callerID, "merchant", &orgID)
	req := httptest.NewRequest(http.MethodGet, "/users/"+otherID.String(), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("merchant accessing another user should get 403, got %d", rec.Code)
	}
}

func TestSelfOrAdmin_InvalidUUIDReturnsBadRequest(t *testing.T) {
	callerID := uuid.New()

	e := newTestEchoForUsers(callerID, "user", nil)
	req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid UUID should return 400, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// enforceSelfOrAdmin — tested via UpdateUserHandler (PUT)
// ---------------------------------------------------------------------------

func TestUpdateUser_SelfPasses(t *testing.T) {
	callerID := uuid.New()

	e := newTestEchoForUsers(callerID, "user", nil)
	req := httptest.NewRequest(http.MethodPut, "/users/"+callerID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("caller updating own record should pass self check, got 403")
	}
}

func TestUpdateUser_OtherUserForbidden(t *testing.T) {
	callerID := uuid.New()
	otherID := uuid.New()

	e := newTestEchoForUsers(callerID, "user", nil)
	req := httptest.NewRequest(http.MethodPut, "/users/"+otherID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin updating another user should get 403, got %d", rec.Code)
	}
}

func TestUpdateUser_AdminBypasses(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()

	e := newTestEchoForUsers(adminID, "administrator", nil)
	req := httptest.NewRequest(http.MethodPut, "/users/"+targetID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("admin should bypass self-or-admin check on update, got 403")
	}
}

// ---------------------------------------------------------------------------
// enforceOrgOwnership — tested via CreateStationHandler
// The CreateStationHandler checks org ownership for non-admin callers:
// if role != "administrator", it requires caller's org_id to be non-nil,
// then overrides req.OrgID with the caller's org. If org_id is nil, 403.
// ---------------------------------------------------------------------------

func newTestEchoForStationCreate(callerID uuid.UUID, role string, orgID *uuid.UUID) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = apperror.HTTPErrorHandler
	e.Validator = &testValidator{v: validator.New()}
	e.Use(echomw.Recover())
	cfg := config.Load()

	g := e.Group("", fakeAuthMiddleware(callerID, role, orgID))
	g.POST("/stations", handler.CreateStationHandler(nil, cfg))

	return e
}

// stationCreateBody returns a valid JSON body for CreateStationRequest.
func stationCreateBody(orgID uuid.UUID) string {
	return fmt.Sprintf(`{"org_id":"%s","name":"Test Station","timezone":"UTC"}`, orgID.String())
}

func TestCreateStation_MerchantWithoutOrgForbidden(t *testing.T) {
	callerID := uuid.New()
	fakeOrgID := uuid.New()

	e := newTestEchoForStationCreate(callerID, "merchant", nil)
	body := stationCreateBody(fakeOrgID)
	req := httptest.NewRequest(http.MethodPost, "/stations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("merchant without org_id creating station should get 403, got %d", rec.Code)
	}
}

func TestCreateStation_AdminWithoutOrgDoesNotGet403(t *testing.T) {
	adminID := uuid.New()
	fakeOrgID := uuid.New()

	e := newTestEchoForStationCreate(adminID, "administrator", nil)
	body := stationCreateBody(fakeOrgID)
	req := httptest.NewRequest(http.MethodPost, "/stations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Admin bypasses the org check. The request may fail for other reasons
	// (nil DB) but should NOT be 403.
	if rec.Code == http.StatusForbidden {
		t.Fatalf("admin should bypass org ownership check, got 403")
	}
}

func TestCreateStation_MerchantWithOrgDoesNotGet403(t *testing.T) {
	callerID := uuid.New()
	orgID := uuid.New()

	e := newTestEchoForStationCreate(callerID, "merchant", &orgID)
	body := stationCreateBody(orgID)
	req := httptest.NewRequest(http.MethodPost, "/stations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Merchant with org passes the org check. Will fail at DB layer (500),
	// but should NOT be 403.
	if rec.Code == http.StatusForbidden {
		t.Fatalf("merchant with org_id should pass org check, got 403")
	}
}

// ---------------------------------------------------------------------------
// Verify that ErrForbidden has the expected HTTP status code and message.
// This ensures the auth helpers return the correct sentinel error.
// ---------------------------------------------------------------------------

func TestErrForbiddenContract(t *testing.T) {
	err := apperror.ErrForbidden

	if err.Code != http.StatusForbidden {
		t.Fatalf("ErrForbidden.Code = %d, want %d", err.Code, http.StatusForbidden)
	}
	if err.Message != "forbidden" {
		t.Fatalf("ErrForbidden.Message = %q, want %q", err.Message, "forbidden")
	}

	// Verify it satisfies the error interface
	var e error = err
	if e.Error() != "forbidden" {
		t.Fatalf("ErrForbidden.Error() = %q, want %q", e.Error(), "forbidden")
	}
}

func TestErrForbiddenRendersAs403InResponse(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = apperror.HTTPErrorHandler

	e.GET("/test", func(c echo.Context) error {
		return apperror.ErrForbidden
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Verify that ErrForbidden is detectable via errors.Is / type assertion.
// This ensures handlers returning ErrForbidden are caught by the error handler.
// ---------------------------------------------------------------------------

func TestErrForbiddenIsAppError(t *testing.T) {
	var appErr *apperror.AppError
	if !errors.As(apperror.ErrForbidden, &appErr) {
		t.Fatal("ErrForbidden should be detectable as *apperror.AppError")
	}
	if appErr.Code != http.StatusForbidden {
		t.Fatalf("detected AppError code = %d, want %d", appErr.Code, http.StatusForbidden)
	}
}
