package API_tests

import (
	"net/http"
	"testing"
)

func TestUserCannotCreateWarehouse(t *testing.T) {
	email := uniqueEmail("wh_create")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/warehouses", map[string]string{
		"name":    "Test Warehouse",
		"address": "456 Warehouse Ave",
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

func TestUserCannotListWarehouses(t *testing.T) {
	email := uniqueEmail("wh_list")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/warehouses", nil, token)
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
