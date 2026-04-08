package API_tests

import (
	"net/http"
	"testing"
)

func TestPricingLifecycle(t *testing.T) {
	_, adminToken := createAdminUser(t, "price_lc")

	// 1. Create org + station
	orgID := adminCreateOrg(t, adminToken, "priceorg")

	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id":   orgID,
		"name":     "Pricing Station",
		"timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create station, got %d: %v", resp.StatusCode, stData)
	}
	stationID := stData["id"].(string)

	// 2. Create price template
	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name":       "Peak Rate",
		"station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create template, got %d: %v", resp.StatusCode, tmplData)
	}
	tmplID := tmplData["id"].(string)

	// 3. Create version
	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate":   "0.25",
		"duration_rate": "0.05",
		"service_fee":   "1.50",
		"apply_tax":     true,
	}, adminToken)
	verData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create version, got %d: %v", resp.StatusCode, verData)
	}
	versionID := verData["id"].(string)

	// 4. List versions and verify version_number is 1
	resp = doRequest(t, "GET", "/api/v1/pricing/templates/"+tmplID+"/versions", nil, adminToken)
	versions := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list versions, got %d", resp.StatusCode)
	}
	if len(versions) < 1 {
		t.Fatal("expected at least 1 version")
	}
	firstVer := versions[0].(map[string]interface{})
	verNum := firstVer["version_number"].(float64)
	if int(verNum) != 1 {
		t.Fatalf("expected version_number 1, got %v", verNum)
	}

	// 5. Add TOU rule
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/tou-rules", map[string]interface{}{
		"day_type":      "weekday",
		"start_time":    "07:00",
		"end_time":      "10:00",
		"energy_rate":   "0.35",
		"duration_rate": "0.08",
	}, adminToken)
	touData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create TOU rule, got %d: %v", resp.StatusCode, touData)
	}

	// 6. List TOU rules and verify 1 rule
	resp = doRequest(t, "GET", "/api/v1/pricing/versions/"+versionID+"/tou-rules", nil, adminToken)
	touRules := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list TOU rules, got %d", resp.StatusCode)
	}
	if len(touRules) != 1 {
		t.Fatalf("expected 1 TOU rule, got %d", len(touRules))
	}

	// 7. Add overlapping TOU rule -> expect 400
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/tou-rules", map[string]interface{}{
		"day_type":      "weekday",
		"start_time":    "08:00",
		"end_time":      "11:00",
		"energy_rate":   "0.40",
		"duration_rate": "0.10",
	}, adminToken)
	overlapData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for overlapping TOU rule, got %d: %v", resp.StatusCode, overlapData)
	}

	// 8. Activate version
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/activate", nil, adminToken)
	actData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on activate, got %d: %v", resp.StatusCode, actData)
	}
	if actData["status"] != "active" {
		t.Fatalf("expected status active, got %v", actData["status"])
	}

	// 9. Deactivate
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/deactivate", nil, adminToken)
	deactData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on deactivate, got %d: %v", resp.StatusCode, deactData)
	}

	// 10. Rollback -> creates new version with incremented number
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/rollback", nil, adminToken)
	rollData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on rollback, got %d: %v", resp.StatusCode, rollData)
	}
	newVerNum := rollData["version_number"].(float64)
	if int(newVerNum) <= int(verNum) {
		t.Fatalf("expected rollback version number > %d, got %v", int(verNum), newVerNum)
	}
}
