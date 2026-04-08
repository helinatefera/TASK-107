package API_tests

import (
	"net/http"
	"testing"
)

func TestCreateVersion_NegativeEnergyRate_Rejected(t *testing.T) {
	_, adminToken := createAdminUser(t, "neg_energy")
	orgID := adminCreateOrg(t, adminToken, "negorg")

	// Create station
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "NegStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create station failed: %d %v", resp.StatusCode, stData)
	}
	stationID := stData["id"].(string)

	// Create template
	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "NegTest", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create template failed: %d %v", resp.StatusCode, tmplData)
	}
	tmplID := tmplData["id"].(string)

	// Try to create version with negative energy_rate
	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate":   "-0.25",
		"duration_rate": "0.05",
		"service_fee":   "1.50",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative energy_rate, got %d: %v", resp.StatusCode, data)
	}
}

func TestCreateVersion_NegativeServiceFee_Rejected(t *testing.T) {
	_, adminToken := createAdminUser(t, "neg_fee")
	orgID := adminCreateOrg(t, adminToken, "negfeeorg")

	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "FeeStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "FeeTest", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate":   "0.25",
		"duration_rate": "0.05",
		"service_fee":   "-1.50",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative service_fee, got %d: %v", resp.StatusCode, data)
	}
}

func TestCreateTOURule_NegativeRate_Rejected(t *testing.T) {
	_, adminToken := createAdminUser(t, "neg_tou")
	orgID := adminCreateOrg(t, adminToken, "negtouorg")

	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "TOUStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "TOUNeg", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.25", "duration_rate": "0.05", "service_fee": "1.00",
	}, adminToken)
	verData := parseJSON(t, resp)
	versionID := verData["id"].(string)

	// Try TOU rule with negative energy_rate
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/tou-rules", map[string]interface{}{
		"day_type": "weekday", "start_time": "08:00", "end_time": "12:00",
		"energy_rate": "-0.10", "duration_rate": "0.05",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative TOU energy_rate, got %d: %v", resp.StatusCode, data)
	}
}

func TestCreateTOURule_OverlapRejected(t *testing.T) {
	_, adminToken := createAdminUser(t, "tou_overlap")
	orgID := adminCreateOrg(t, adminToken, "overlaporg")

	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "OverlapStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "OverlapTest", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.25", "duration_rate": "0.05", "service_fee": "1.00",
	}, adminToken)
	verData := parseJSON(t, resp)
	versionID := verData["id"].(string)

	// First rule: 08:00-12:00
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/tou-rules", map[string]interface{}{
		"day_type": "weekday", "start_time": "08:00", "end_time": "12:00",
		"energy_rate": "0.30", "duration_rate": "0.08",
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		data := parseJSON(t, resp)
		t.Fatalf("first TOU rule failed: %d %v", resp.StatusCode, data)
	}

	// Overlapping rule: 10:00-14:00 (overlaps 10:00-12:00)
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/tou-rules", map[string]interface{}{
		"day_type": "weekday", "start_time": "10:00", "end_time": "14:00",
		"energy_rate": "0.35", "duration_rate": "0.10",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for overlapping TOU rule, got %d: %v", resp.StatusCode, data)
	}
}
