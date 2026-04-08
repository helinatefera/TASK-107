package API_tests

import (
	"net/http"
	"testing"
)

// registerAndGetID registers a user and returns their ID from the register response.
func registerAndGetID(t *testing.T, email string) string {
	t.Helper()
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on register, got %d: %v", resp.StatusCode, data)
	}
	id, ok := data["id"].(string)
	if !ok {
		t.Fatal("expected id field in register response")
	}
	return id
}

// loginOnly logs in an already-registered user and returns a bearer token.
func loginOnly(t *testing.T, email string) string {
	t.Helper()
	resp := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":     email,
		"password":  "SecurePass123!",
		"device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d: %v", resp.StatusCode, data)
	}
	token, ok := data["token"].(string)
	if !ok {
		t.Fatal("expected token in login response")
	}
	return token
}

func TestUserCannotAccessOtherUserProfile(t *testing.T) {
	emailA := uniqueEmail("iso_a_profile")
	emailB := uniqueEmail("iso_b_profile")

	// Register both users, capture B's ID
	_ = registerAndGetID(t, emailA)
	idB := registerAndGetID(t, emailB)

	// Login as user A
	tokenA := loginOnly(t, emailA)

	// User A tries to access user B's profile
	resp := doRequest(t, "GET", "/api/v1/users/"+idB, nil, tokenA)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %v", resp.StatusCode, data)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected error code 403, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestUserCannotUpdateOtherUserProfile(t *testing.T) {
	emailA := uniqueEmail("iso_a_update")
	emailB := uniqueEmail("iso_b_update")

	_ = registerAndGetID(t, emailA)
	idB := registerAndGetID(t, emailB)

	tokenA := loginOnly(t, emailA)

	resp := doRequest(t, "PUT", "/api/v1/users/"+idB, map[string]string{
		"display_name": "Hacked Name",
	}, tokenA)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %v", resp.StatusCode, data)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected error code 403, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestUserCannotGetOtherUserPermissions(t *testing.T) {
	emailA := uniqueEmail("iso_a_perm")
	emailB := uniqueEmail("iso_b_perm")

	_ = registerAndGetID(t, emailA)
	idB := registerAndGetID(t, emailB)

	tokenA := loginOnly(t, emailA)

	resp := doRequest(t, "GET", "/api/v1/users/"+idB+"/permissions", nil, tokenA)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %v", resp.StatusCode, data)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected error code 403, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}
