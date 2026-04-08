package API_tests

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAdminCanViewConfig(t *testing.T) {
	_, token := createAdminUser(t, "adm_cfg_view")

	resp := doRequest(t, "GET", "/api/v1/admin/config", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
}

func TestAdminCanUpdateConfig(t *testing.T) {
	_, token := createAdminUser(t, "adm_cfg_upd")

	// Update tax_rate
	resp := doRequest(t, "PUT", "/api/v1/admin/config/tax_rate", map[string]interface{}{
		"value": "0.10",
	}, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, data)
	}

	// Verify the update by reading config
	resp2 := doRequest(t, "GET", "/api/v1/admin/config", nil, token)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on GET config, got %d", resp2.StatusCode)
	}
	body, _ := io.ReadAll(resp2.Body)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "tax_rate") {
		t.Fatal("expected config to contain tax_rate")
	}
	if !strings.Contains(bodyStr, "0.10") {
		t.Fatal("expected tax_rate value to be 0.10")
	}
}

func TestAdminCanViewAuditLogs(t *testing.T) {
	_, token := createAdminUser(t, "adm_audit")

	resp := doRequest(t, "GET", "/api/v1/admin/audit-logs", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
}

func TestAdminCanViewMetrics(t *testing.T) {
	_, token := createAdminUser(t, "adm_metrics")

	resp := doRequest(t, "GET", "/api/v1/admin/metrics", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
}

func TestAdminCanExportMetricsCSV(t *testing.T) {
	_, token := createAdminUser(t, "adm_csv")

	resp := doRequest(t, "GET", "/api/v1/admin/metrics/export", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Fatalf("expected Content-Type to contain text/csv, got %s", ct)
	}
}

func TestAdminCanExportMetricsJSON(t *testing.T) {
	_, token := createAdminUser(t, "adm_json")

	resp := doRequest(t, "GET", "/api/v1/admin/metrics/export?format=json", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected Content-Type to contain application/json, got %s", ct)
	}
}
