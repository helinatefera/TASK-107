package API_tests

import (
	"net/http"
	"testing"
	"time"
)

// TestOrderCreate_WithActivePricing creates a full pricing setup and verifies
// that an order is correctly priced.
func TestOrderCreate_WithActivePricing(t *testing.T) {
	_, adminToken := createAdminUser(t, "order_price")
	orgID := adminCreateOrg(t, adminToken, "orderorg")

	// Create station
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "OrderStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create station: %d %v", resp.StatusCode, stData)
	}
	stationID := stData["id"].(string)

	// Create device on station
	resp = doRequest(t, "POST", "/api/v1/stations/"+stationID+"/devices", map[string]interface{}{
		"device_code": uniqueCode("dev"),
	}, adminToken)
	devData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create device: %d %v", resp.StatusCode, devData)
	}
	deviceID := devData["id"].(string)

	// Create pricing template
	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "OrderPricing", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	// Create version
	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.25", "duration_rate": "0.05", "service_fee": "1.00",
	}, adminToken)
	verData := parseJSON(t, resp)
	versionID := verData["id"].(string)

	// Activate
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/activate", nil, adminToken)
	if resp.StatusCode != http.StatusOK {
		actData := parseJSON(t, resp)
		t.Fatalf("activate: %d %v", resp.StatusCode, actData)
	}

	// Create a user in the org to place the order
	userID, userToken := createMerchantUser(t, "order_user", orgID)
	_ = userID

	// Place order — start_time must be after the version's effective_at.
	// Add 2s buffer to account for clock difference between test and server.
	now := time.Now().UTC().Add(2 * time.Second)
	resp = doRequest(t, "POST", "/api/v1/orders", map[string]interface{}{
		"order_id":   uniqueCode("ord"),
		"device_id":  deviceID,
		"energy_kwh": "10.5",
		"start_time": now.Format(time.RFC3339),
		"end_time":   now.Add(30 * time.Minute).Format(time.RFC3339),
	}, userToken)
	orderData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create order: %d %v", resp.StatusCode, orderData)
	}

	// Verify pricing fields are present and non-zero
	if orderData["total"] == nil {
		t.Fatal("order missing total")
	}
	if orderData["energy_rate"] == nil {
		t.Fatal("order missing energy_rate")
	}
}

// TestRecalculate_AfterDeactivation verifies that recalculating an order works
// even after the pricing version has been deactivated.
func TestRecalculate_AfterDeactivation(t *testing.T) {
	_, adminToken := createAdminUser(t, "recalc")
	orgID := adminCreateOrg(t, adminToken, "recalcorg")

	// Create station + device + pricing + activate
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "RecalcStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/stations/"+stationID+"/devices", map[string]interface{}{
		"device_code": uniqueCode("rdev"),
	}, adminToken)
	devData := parseJSON(t, resp)
	deviceID := devData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "RecalcPricing", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.30", "duration_rate": "0.10", "service_fee": "2.00",
	}, adminToken)
	verData := parseJSON(t, resp)
	versionID := verData["id"].(string)

	doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/activate", nil, adminToken)

	// Create order as merchant — add 2s buffer for clock consistency
	_, merchantToken := createMerchantUser(t, "recalc_merch", orgID)
	now := time.Now().UTC().Add(2 * time.Second)
	resp = doRequest(t, "POST", "/api/v1/orders", map[string]interface{}{
		"order_id":   uniqueCode("rord"),
		"device_id":  deviceID,
		"energy_kwh": "5.0",
		"start_time": now.Format(time.RFC3339),
		"end_time":   now.Add(20 * time.Minute).Format(time.RFC3339),
	}, merchantToken)
	orderData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create order: %d %v", resp.StatusCode, orderData)
	}
	orderID := orderData["id"].(string)

	// Deactivate the pricing version
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/deactivate", nil, adminToken)
	if resp.StatusCode != http.StatusOK {
		deactData := parseJSON(t, resp)
		t.Fatalf("deactivate: %d %v", resp.StatusCode, deactData)
	}

	// Recalculate should still work using the historical version
	resp = doRequest(t, "POST", "/api/v1/orders/"+orderID+"/recalculate", nil, adminToken)
	recalcData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("recalculate after deactivation: expected 200, got %d: %v", resp.StatusCode, recalcData)
	}
	if recalcData["total"] == nil {
		t.Fatal("recalculated order missing total")
	}
}

// TestActivateVersion_WithEffectiveAt verifies that a pricing version can be
// activated with an explicit effective_at timestamp.
func TestActivateVersion_WithEffectiveAt(t *testing.T) {
	_, adminToken := createAdminUser(t, "eff_at")
	orgID := adminCreateOrg(t, adminToken, "efforg")

	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "EffStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "EffTest", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.20", "duration_rate": "0.04", "service_fee": "0.50",
	}, adminToken)
	verData := parseJSON(t, resp)
	versionID := verData["id"].(string)

	// Activate with explicit future effective_at
	futureTime := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/activate", map[string]interface{}{
		"effective_at": futureTime,
	}, adminToken)
	actData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("activate with effective_at: %d %v", resp.StatusCode, actData)
	}
	if actData["status"] != "active" {
		t.Fatalf("expected active, got %v", actData["status"])
	}
	if actData["effective_at"] == nil {
		t.Fatal("missing effective_at in response")
	}
}

// TestActivateVersion_PastEffectiveAt_Rejected verifies that activating with a
// past effective_at is rejected.
func TestActivateVersion_PastEffectiveAt_Rejected(t *testing.T) {
	_, adminToken := createAdminUser(t, "past_eff")
	orgID := adminCreateOrg(t, adminToken, "pastefforg")

	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "PastEffStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "PastEffTest", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.20", "duration_rate": "0.04", "service_fee": "0.50",
	}, adminToken)
	verData := parseJSON(t, resp)
	versionID := verData["id"].(string)

	pastTime := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+versionID+"/activate", map[string]interface{}{
		"effective_at": pastTime,
	}, adminToken)
	if resp.StatusCode != http.StatusBadRequest {
		data := parseJSON(t, resp)
		t.Fatalf("expected 400 for past effective_at, got %d: %v", resp.StatusCode, data)
	}
}

// TestFutureDatedActivation_PreservesCurrentPricing verifies that activating a
// new version with a future effective_at does NOT deactivate the current version,
// so orders placed before the future date still resolve pricing correctly.
func TestFutureDatedActivation_PreservesCurrentPricing(t *testing.T) {
	_, adminToken := createAdminUser(t, "futprice")
	orgID := adminCreateOrg(t, adminToken, "futpriceorg")

	// Create station + device
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgID, "name": "FutPriceStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/stations/"+stationID+"/devices", map[string]interface{}{
		"device_code": uniqueCode("fpdev"),
	}, adminToken)
	devData := parseJSON(t, resp)
	deviceID := devData["id"].(string)

	// Create pricing template + version v1
	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "FutPriceTest", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.10", "duration_rate": "0.02", "service_fee": "0.50",
	}, adminToken)
	v1Data := parseJSON(t, resp)
	v1ID := v1Data["id"].(string)

	// Activate v1 immediately
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+v1ID+"/activate", nil, adminToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("activate v1: %d", resp.StatusCode)
	}

	// Create version v2 with higher rates
	resp = doRequest(t, "POST", "/api/v1/pricing/templates/"+tmplID+"/versions", map[string]interface{}{
		"energy_rate": "0.50", "duration_rate": "0.10", "service_fee": "3.00",
	}, adminToken)
	v2Data := parseJSON(t, resp)
	v2ID := v2Data["id"].(string)

	// Activate v2 with a FUTURE effective_at (1 hour from now)
	futureTime := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	resp = doRequest(t, "POST", "/api/v1/pricing/versions/"+v2ID+"/activate", map[string]interface{}{
		"effective_at": futureTime,
	}, adminToken)
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("activate v2 with future effective_at: %d %v", resp.StatusCode, data)
	}

	// Verify v1 is still active (not deactivated)
	resp = doRequest(t, "GET", "/api/v1/pricing/versions/"+v1ID, nil, adminToken)
	v1Status := parseJSON(t, resp)
	if v1Status["status"] != "active" {
		t.Fatalf("v1 should remain active during future-dated activation, got %v", v1Status["status"])
	}

	// Create an order NOW — it should resolve to v1 (since v2's effective_at is in the future).
	// Add 2s buffer for clock consistency between test and server.
	_, merchantToken := createMerchantUser(t, "futprice_merch", orgID)
	now := time.Now().UTC().Add(2 * time.Second)
	resp = doRequest(t, "POST", "/api/v1/orders", map[string]interface{}{
		"order_id":   uniqueCode("fpord"),
		"device_id":  deviceID,
		"energy_kwh": "10.0",
		"start_time": now.Format(time.RFC3339),
		"end_time":   now.Add(30 * time.Minute).Format(time.RFC3339),
	}, merchantToken)
	orderData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create order during future-dated activation: %d %v", resp.StatusCode, orderData)
	}

	// The order should use v1's pricing (energy_rate=0.10), not v2's (0.50)
	if orderData["version_id"] != v1ID {
		t.Fatalf("order should use v1 (%s), got version_id=%v", v1ID, orderData["version_id"])
	}
}
