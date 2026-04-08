package API_tests

import (
	"net/http"
	"testing"
)

// TestLockout_AfterMaxFailedAttempts verifies that repeated wrong passwords
// trigger account lockout (10 failed attempts in the lockout window).
func TestLockout_AfterMaxFailedAttempts(t *testing.T) {
	email := uniqueEmail("lockout")
	// Register the user
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email": email, "password": "SecurePass123!", "display_name": "Lockout User",
	}, "")
	if resp.StatusCode != http.StatusCreated {
		data := parseJSON(t, resp)
		t.Fatalf("register failed: %d %v", resp.StatusCode, data)
	}
	resp.Body.Close()

	// Send 10 wrong-password attempts
	for i := 0; i < 10; i++ {
		resp = doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
			"email": email, "password": "WrongPassword1!", "device_id": "test-device",
		}, "")
		resp.Body.Close()
	}

	// The 11th attempt (even with the correct password) should be locked
	resp = doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": "SecurePass123!", "device_id": "test-device",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 (account locked), got %d: %v", resp.StatusCode, data)
	}
}

// TestDeviceMismatch_Rejected verifies that using a token from device A
// with device B's X-Device-Id header is rejected.
func TestDeviceMismatch_Rejected(t *testing.T) {
	email := uniqueEmail("devmismatch")
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email": email, "password": "SecurePass123!", "display_name": "Device User",
	}, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Login with device-A
	resp = doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": "SecurePass123!", "device_id": "device-A",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: %d %v", resp.StatusCode, data)
	}
	token := data["token"].(string)

	// Use that token but with a different X-Device-Id header
	req, _ := http.NewRequest("GET", baseURL()+"/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Device-Id", "device-B")
	req.Header.Set("Content-Type", "application/json")
	devResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer devResp.Body.Close()

	// Should be rejected (device mismatch)
	if devResp.StatusCode != http.StatusUnauthorized && devResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 401 or 403 for device mismatch, got %d", devResp.StatusCode)
	}
}

// TestRefresh_ValidSession verifies that refresh works on a valid session.
func TestRefresh_ValidSession(t *testing.T) {
	email := uniqueEmail("refresh_valid")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/auth/refresh", map[string]string{
		"device_id": "test-device-001",
	}, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on refresh, got %d: %v", resp.StatusCode, data)
	}
	if data["status"] != "refreshed" {
		t.Fatalf("expected status=refreshed, got %v", data["status"])
	}
}

// TestRefresh_WrongDevice verifies refresh is rejected with wrong device.
func TestRefresh_WrongDevice(t *testing.T) {
	email := uniqueEmail("refresh_dev")
	token := registerAndLogin(t, email)

	// Try refresh with wrong device_id
	resp := doRequest(t, "POST", "/api/v1/auth/refresh", map[string]string{
		"device_id": "wrong-device-999",
	}, token)
	data := parseJSON(t, resp)
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected error for wrong device, got 200: %v", data)
	}
}

// TestRecovery_NonExistentEmail_NoLeak verifies recovery returns a token-shaped
// response even for non-existent emails (no account enumeration).
func TestRecovery_NonExistentEmail_NoLeak(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/auth/recover", map[string]string{
		"email": "absolutely_not_a_user@nowhere.test",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (no leak), got %d: %v", resp.StatusCode, data)
	}
	token, ok := data["token"].(string)
	if !ok || len(token) < 32 {
		t.Fatalf("expected token-shaped response for non-existent email, got: %v", data)
	}
}
