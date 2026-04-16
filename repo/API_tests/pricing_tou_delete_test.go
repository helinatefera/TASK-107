package API_tests

import (
	"net/http"
	"testing"
)

// TestDeleteTOURule covers DELETE /api/v1/pricing/tou-rules/:id.
func TestDeleteTOURule(t *testing.T) {
	_, adminToken := createAdminUser(t, "tou_del")

	// Setup: org → station → template → version → TOU rule
	orgID := adminCreateOrg(t, adminToken, "tou_del_org")

	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id":   orgID,
		"name":     "TOU Station",
		"timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create station: %d %v", resp.StatusCode, stData)
	}
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name":       "TOU Template",
		"station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create template: %d %v", resp.StatusCode, tmplData)
	}
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate":   "0.20",
		"duration_rate": "0.04",
		"service_fee":   "1.00",
		"apply_tax":     false,
	}, adminToken)
	verData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create version: %d %v", resp.StatusCode, verData)
	}
	versionID := verData["id"].(string)

	// Create a TOU rule
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/tou-rules", map[string]interface{}{
		"day_type":      "weekday",
		"start_time":    "06:00",
		"end_time":      "12:00",
		"energy_rate":   "0.30",
		"duration_rate": "0.06",
	}, adminToken)
	touData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create TOU rule: %d %v", resp.StatusCode, touData)
	}
	touRuleID := touData["id"].(string)

	// Verify rule exists in list
	resp = doRequest(t, "GET", "/api/v1/pricing/versions/"+versionID+"/tou-rules", nil, adminToken)
	rules := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list TOU rules: %d", resp.StatusCode)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 TOU rule, got %d", len(rules))
	}

	// Delete the TOU rule
	resp = doRequest(t, "DELETE", "/api/v1/pricing/tou-rules/"+touRuleID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete TOU rule, got %d", resp.StatusCode)
	}

	// Confirm rule is gone
	resp = doRequest(t, "GET", "/api/v1/pricing/versions/"+versionID+"/tou-rules", nil, adminToken)
	rules = parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list TOU rules after delete: %d", resp.StatusCode)
	}
	if len(rules) != 0 {
		t.Fatalf("expected 0 TOU rules after delete, got %d", len(rules))
	}
}

// TestMerchantCanDeleteOwnOrgTOURule verifies merchant can delete TOU rules for their org.
func TestMerchantCanDeleteOwnOrgTOURule(t *testing.T) {
	_, adminToken := createAdminUser(t, "tou_merch_del")
	orgID := adminCreateOrg(t, adminToken, "tou_merch_org")
	_, merchantToken := createMerchantUser(t, "tou_merch", orgID)

	// Admin creates station + template + version
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id":   orgID,
		"name":     "Merch TOU Station",
		"timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name":       "Merch TOU Tmpl",
		"station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate":   "0.15",
		"duration_rate": "0.03",
		"service_fee":   "0.50",
		"apply_tax":     false,
	}, adminToken)
	verData := parseJSON(t, resp)
	versionID := verData["id"].(string)

	// Merchant creates a TOU rule
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/tou-rules", map[string]interface{}{
		"day_type":      "weekend",
		"start_time":    "08:00",
		"end_time":      "20:00",
		"energy_rate":   "0.10",
		"duration_rate": "0.02",
	}, merchantToken)
	touData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("merchant create TOU rule: %d %v", resp.StatusCode, touData)
	}
	touRuleID := touData["id"].(string)

	// Merchant deletes the TOU rule
	resp = doRequest(t, "DELETE", "/api/v1/pricing/tou-rules/"+touRuleID, nil, merchantToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on merchant delete TOU rule, got %d", resp.StatusCode)
	}
}
