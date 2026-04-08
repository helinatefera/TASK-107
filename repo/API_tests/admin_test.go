package API_tests

import (
	"net/http"
	"testing"
)

func TestNonAdminCannotListAuditLogs(t *testing.T) {
	email := uniqueEmail("admin_audit")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/admin/audit-logs", nil, token)
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

func TestNonAdminCannotViewConfig(t *testing.T) {
	email := uniqueEmail("admin_config_view")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/admin/config", nil, token)
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

func TestNonAdminCannotUpdateConfig(t *testing.T) {
	email := uniqueEmail("admin_config_upd")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "PUT", "/api/v1/admin/config/tax_rate", map[string]interface{}{
		"value": "0.08",
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

func TestNonAdminCannotViewMetrics(t *testing.T) {
	email := uniqueEmail("admin_metrics")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/admin/metrics", nil, token)
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

func TestNonAdminCannotExportMetrics(t *testing.T) {
	email := uniqueEmail("admin_export")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/admin/metrics/export", nil, token)
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
