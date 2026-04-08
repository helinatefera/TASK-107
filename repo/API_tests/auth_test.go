package API_tests

import (
	"net/http"
	"testing"
)

func TestRegister_ValidData(t *testing.T) {
	email := uniqueEmail("reg_valid")
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	email := uniqueEmail("reg_short")
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "Ab1!",
		"display_name": "Test User",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected error code 400, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestRegister_WeakPassword_TwoClasses(t *testing.T) {
	email := uniqueEmail("reg_weak")
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "alllowercase",
		"display_name": "Test User",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected error code 400, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestRegister_MissingEmail(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected error code 400, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	email := uniqueEmail("reg_dup")
	// First registration
	resp1 := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first register expected 201, got %d", resp1.StatusCode)
	}

	// Second registration with same email
	resp2 := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User 2",
	}, "")
	data := parseJSON(t, resp2)
	if resp2.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", resp2.StatusCode)
	}
	if _, ok := data["code"].(float64); !ok {
		t.Fatal("expected code field in error response")
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestLogin_ValidCredentials(t *testing.T) {
	email := uniqueEmail("login_valid")
	// Register first
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	resp.Body.Close()

	// Login
	resp = doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":     email,
		"password":  "SecurePass123!",
		"device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if _, ok := data["token"].(string); !ok {
		t.Fatal("expected token in login response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	email := uniqueEmail("login_wrong")
	// Register first
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	resp.Body.Close()

	// Login with wrong password
	resp = doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":     email,
		"password":  "WrongPassword999!",
		"device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 401 {
		t.Fatalf("expected error code 401, got %v", data["code"])
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	email := uniqueEmail("login_noexist")
	resp := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":     email,
		"password":  "SecurePass123!",
		"device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 401 {
		t.Fatalf("expected error code 401, got %v", data["code"])
	}
}

func TestLogin_MissingDeviceID(t *testing.T) {
	email := uniqueEmail("login_nodev")
	// Register first
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	resp.Body.Close()

	// Login without device_id
	resp = doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "SecurePass123!",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected error code 400, got %v", data["code"])
	}
}

func TestLogout_ValidToken(t *testing.T) {
	email := uniqueEmail("logout_valid")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/auth/logout", nil, token)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestRefresh_ValidToken(t *testing.T) {
	email := uniqueEmail("refresh_valid")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/auth/refresh", map[string]string{
		"device_id": "test-device-001",
	}, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, data)
	}
	if status, ok := data["status"].(string); !ok || status != "refreshed" {
		t.Fatal("expected status=refreshed in refresh response")
	}
}

func TestRecover_ValidEmail(t *testing.T) {
	email := uniqueEmail("recover_valid")
	// Register first
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	resp.Body.Close()

	// Request recovery
	resp = doRequest(t, "POST", "/api/v1/auth/recover", map[string]string{
		"email": email,
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if _, ok := data["token"].(string); !ok {
		t.Fatal("expected token in recovery response")
	}
}

func TestRecoverReset_ValidTokenAndPassword(t *testing.T) {
	email := uniqueEmail("recover_reset")
	// Register
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	resp.Body.Close()

	// Get recovery token
	resp = doRequest(t, "POST", "/api/v1/auth/recover", map[string]string{
		"email": email,
	}, "")
	data := parseJSON(t, resp)
	recoveryToken, ok := data["token"].(string)
	if !ok {
		t.Fatal("failed to get recovery token")
	}

	// Reset password
	resp = doRequest(t, "POST", "/api/v1/auth/recover/reset", map[string]string{
		"token":        recoveryToken,
		"new_password": "NewSecurePass456!",
	}, "")
	if resp.StatusCode != http.StatusOK {
		body := parseJSON(t, resp)
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, body)
	}
	resp.Body.Close()
}

func TestErrorFormat_Consistency(t *testing.T) {
	// Test several error endpoints to verify consistent {"code":N,"msg":"..."} format
	tests := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{
			name:   "register missing fields",
			method: "POST",
			path:   "/api/v1/auth/register",
			body:   map[string]string{},
		},
		{
			name:   "login missing fields",
			method: "POST",
			path:   "/api/v1/auth/login",
			body:   map[string]string{},
		},
		{
			name:   "login wrong credentials",
			method: "POST",
			path:   "/api/v1/auth/login",
			body: map[string]string{
				"email":     "nonexistent@test.com",
				"password":  "SomePass123!",
				"device_id": "test-device-001",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, tc.method, tc.path, tc.body, "")
			data := parseJSON(t, resp)

			if _, ok := data["code"]; !ok {
				t.Error("error response missing 'code' field")
			}
			if _, ok := data["msg"]; !ok {
				t.Error("error response missing 'msg' field")
			}
		})
	}
}
