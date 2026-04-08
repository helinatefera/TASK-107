package API_tests

import (
	"net/http"
	"testing"
)

func TestNonAdminCannotDeleteStation(t *testing.T) {
	email := uniqueEmail("station_del")
	token := registerAndLogin(t, email)

	// Regular user (role=user) tries to create a station — needs merchant/admin
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]string{
		"name":    "Test Station",
		"address": "123 Test St",
	}, token)
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

func TestUserCannotAccessStations(t *testing.T) {
	email := uniqueEmail("station_list")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/stations", nil, token)
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

func TestUserCannotAccessDevices(t *testing.T) {
	email := uniqueEmail("device_get")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "PUT", "/api/v1/devices/00000000-0000-0000-0000-000000000001", map[string]string{
		"device_code": "TEST",
		"device_type": "AC",
		"status":      "active",
	}, token)
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
