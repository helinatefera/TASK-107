package API_tests

import (
	"net/http"
	"testing"
)

func TestUserCannotAccessPricingTemplates(t *testing.T) {
	email := uniqueEmail("price_list")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/pricing/templates", nil, token)
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

func TestUserCannotCreatePricingTemplate(t *testing.T) {
	email := uniqueEmail("price_create")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name":        "Test Template",
		"base_rate":   0.25,
		"per_kwh":     0.10,
		"currency":    "USD",
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
