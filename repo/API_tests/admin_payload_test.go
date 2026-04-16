package API_tests

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestAuditLogPayloadSemantics verifies audit log entries contain expected fields
// and that a config update generates an audit trail with old/new values.
func TestAuditLogPayloadSemantics(t *testing.T) {
	_, adminToken := createAdminUser(t, "audit_payload")

	// Perform an audited action: update a config key
	resp := doRequest(t, "PUT", "/api/v1/admin/config/tax_rate", map[string]interface{}{
		"value": "0.07",
	}, adminToken)
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("expected 200 on config update, got %d: %v", resp.StatusCode, data)
	}
	resp.Body.Close()

	// Fetch audit logs filtered by entity_type=tax_rate
	// (extractAuditEntity uses the last path segment: /admin/config/tax_rate → "tax_rate")
	resp = doRequest(t, "GET", "/api/v1/admin/audit-logs?entity_type=tax_rate&limit=10", nil, adminToken)
	logs := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on audit-logs, got %d", resp.StatusCode)
	}

	// Verify at least one audit entry has the expected structure
	found := false
	for _, entry := range logs {
		m := entry.(map[string]interface{})
		action, _ := m["action"].(string)
		if !strings.Contains(action, "PUT") {
			continue
		}
		found = true

		// Required fields on every audit log entry
		if _, ok := m["id"].(float64); !ok {
			t.Fatal("audit log missing 'id' field")
		}
		if _, ok := m["user_id"].(string); !ok {
			t.Fatal("audit log missing 'user_id' field")
		}
		if _, ok := m["action"].(string); !ok {
			t.Fatal("audit log missing 'action' field")
		}
		if _, ok := m["entity_type"].(string); !ok {
			t.Fatal("audit log missing 'entity_type' field")
		}
		if _, ok := m["created_at"].(string); !ok {
			t.Fatal("audit log missing 'created_at' field")
		}

		// new_value should be present for config updates
		if m["new_value"] == nil {
			t.Fatal("expected new_value to be populated for config update audit entry")
		}
		break
	}
	if !found {
		t.Fatal("no PUT audit log entry found for config update")
	}
}

// TestAuditLogCapturesOrgCreate verifies creating an org produces an audit entry
// with the new org's entity_id.
func TestAuditLogCapturesOrgCreate(t *testing.T) {
	_, adminToken := createAdminUser(t, "audit_org")

	orgCode := uniqueCode("AUDITORG")
	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code": orgCode,
		"name":     "Audit Org",
		"timezone": "UTC",
	}, adminToken)
	orgData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create org: %d %v", resp.StatusCode, orgData)
	}
	orgID := orgData["id"].(string)

	// Fetch audit logs for orgs
	resp = doRequest(t, "GET", "/api/v1/admin/audit-logs?entity_type=orgs&limit=20", nil, adminToken)
	logs := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("audit-logs: %d", resp.StatusCode)
	}

	found := false
	for _, entry := range logs {
		m := entry.(map[string]interface{})
		if m["entity_id"] == orgID {
			found = true
			action, _ := m["action"].(string)
			if !strings.Contains(action, "POST") {
				t.Fatalf("expected POST action for org create, got %q", action)
			}
			if m["new_value"] == nil {
				t.Fatal("expected new_value in org create audit entry")
			}
			break
		}
	}
	if !found {
		t.Fatalf("audit log entry with entity_id=%s not found", orgID)
	}
}

// TestMetricsExportCSVPayload verifies the CSV export contains actual metric rows
// with expected column headers.
func TestMetricsExportCSVPayload(t *testing.T) {
	_, adminToken := createAdminUser(t, "metrics_csv_payload")

	// Make a few requests to generate metrics
	doRequest(t, "GET", "/health", nil, "").Body.Close()
	doRequest(t, "GET", "/health", nil, "").Body.Close()

	resp := doRequest(t, "GET", "/api/v1/admin/metrics/export", nil, adminToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Fatalf("expected text/csv, got %s", ct)
	}

	// Read body and check header row
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(bodyBytes)

	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected CSV header + at least 1 data row, got %d lines", len(lines))
	}

	header := lines[0]
	expectedCols := []string{"id", "method", "path", "status_code", "latency_ms", "recorded_at"}
	for _, col := range expectedCols {
		if !strings.Contains(header, col) {
			t.Fatalf("CSV header missing column %q: %s", col, header)
		}
	}

	// Verify a data row has the expected number of fields
	dataCols := strings.Split(lines[1], ",")
	if len(dataCols) != 6 {
		t.Fatalf("expected 6 columns in data row, got %d: %s", len(dataCols), lines[1])
	}
}

// TestMetricsExportJSONPayload verifies the JSON export contains metric objects
// with expected fields.
func TestMetricsExportJSONPayload(t *testing.T) {
	_, adminToken := createAdminUser(t, "metrics_json_payload")

	resp := doRequest(t, "GET", "/api/v1/admin/metrics/export?format=json", nil, adminToken)
	metrics := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if len(metrics) == 0 {
		t.Fatal("expected at least 1 metric in JSON export")
	}

	first := metrics[0].(map[string]interface{})
	requiredFields := []string{"id", "method", "path", "status_code", "latency_ms", "recorded_at"}
	for _, f := range requiredFields {
		if _, ok := first[f]; !ok {
			t.Fatalf("metric object missing field %q", f)
		}
	}
}

// TestMetricsListPayloadSemantics verifies GET /admin/metrics returns objects
// with proper field types.
func TestMetricsListPayloadSemantics(t *testing.T) {
	_, adminToken := createAdminUser(t, "metrics_list_payload")

	resp := doRequest(t, "GET", "/api/v1/admin/metrics?limit=5", nil, adminToken)
	metrics := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(metrics) == 0 {
		t.Fatal("expected at least 1 metric")
	}

	m := metrics[0].(map[string]interface{})
	if _, ok := m["method"].(string); !ok {
		t.Fatalf("expected method to be string, got %T", m["method"])
	}
	if _, ok := m["path"].(string); !ok {
		t.Fatalf("expected path to be string, got %T", m["path"])
	}
	if _, ok := m["status_code"].(float64); !ok {
		t.Fatalf("expected status_code to be number, got %T", m["status_code"])
	}
	if _, ok := m["latency_ms"].(float64); !ok {
		t.Fatalf("expected latency_ms to be number, got %T", m["latency_ms"])
	}
}
