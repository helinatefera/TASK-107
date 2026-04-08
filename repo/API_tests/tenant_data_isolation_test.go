package API_tests

import (
	"net/http"
	"testing"
)

func TestMerchantCanOnlyAccessOwnOrgStations(t *testing.T) {
	_, adminToken := createAdminUser(t, "iso_st")

	// 1. Create Org A and Org B
	orgAID := adminCreateOrg(t, adminToken, "isoA_st")
	orgBID := adminCreateOrg(t, adminToken, "isoB_st")

	// 2. Admin creates station in Org A
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id":   orgAID,
		"name":     "Org A Station",
		"timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %v", resp.StatusCode, stData)
	}
	stationID := stData["id"].(string)

	// 3. Merchant in Org A can access the station
	_, merchantAToken := createMerchantUser(t, "merch_a_st", orgAID)
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, merchantAToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected merchant A to access station (200), got %d: %v", resp.StatusCode, data)
	}

	// 4. Merchant in Org B cannot access the station
	_, merchantBToken := createMerchantUser(t, "merch_b_st", orgBID)
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, merchantBToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected merchant B to get 403, got %d: %v", resp.StatusCode, data)
	}
}

func TestMerchantCanOnlyAccessOwnOrgWarehouses(t *testing.T) {
	_, adminToken := createAdminUser(t, "iso_wh")

	// 1. Create Org A and Org B
	orgAID := adminCreateOrg(t, adminToken, "isoA_wh")
	orgBID := adminCreateOrg(t, adminToken, "isoB_wh")

	// 2. Admin creates warehouse in Org A
	resp := doRequest(t, "POST", "/api/v1/warehouses", map[string]interface{}{
		"org_id":   orgAID,
		"name":     "Org A Warehouse",
		"timezone": "UTC",
	}, adminToken)
	whData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %v", resp.StatusCode, whData)
	}
	warehouseID := whData["id"].(string)

	// 3. Merchant in Org A can access the warehouse
	_, merchantAToken := createMerchantUser(t, "merch_a_wh", orgAID)
	resp = doRequest(t, "GET", "/api/v1/warehouses/"+warehouseID, nil, merchantAToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected merchant A to access warehouse (200), got %d: %v", resp.StatusCode, data)
	}

	// 4. Merchant in Org B cannot access the warehouse
	_, merchantBToken := createMerchantUser(t, "merch_b_wh", orgBID)
	resp = doRequest(t, "GET", "/api/v1/warehouses/"+warehouseID, nil, merchantBToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected merchant B to get 403, got %d: %v", resp.StatusCode, data)
	}
}

func TestMerchantCreateStationForcesOwnOrg(t *testing.T) {
	_, adminToken := createAdminUser(t, "iso_force")

	// 1. Create Org A and Org B
	orgAID := adminCreateOrg(t, adminToken, "forceA")
	orgBID := adminCreateOrg(t, adminToken, "forceB")

	// 2. Merchant in Org A tries to create station with org_id = Org B
	_, merchantAToken := createMerchantUser(t, "merch_force", orgAID)
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id":   orgBID,
		"name":     "Sneaky Station",
		"timezone": "UTC",
	}, merchantAToken)
	data := parseJSON(t, resp)

	// The API should either reject (403) or force the station to Org A
	if resp.StatusCode == http.StatusCreated {
		// Station was created - verify it belongs to Org A, not Org B
		stationID := data["id"].(string)
		resp2 := doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, adminToken)
		stData := parseJSON(t, resp2)
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 on admin GET station, got %d", resp2.StatusCode)
		}
		if stData["org_id"] != orgAID {
			t.Fatalf("expected station org_id to be forced to %s (Org A), got %v", orgAID, stData["org_id"])
		}
	} else if resp.StatusCode == http.StatusForbidden {
		// Also acceptable: the API rejects the cross-org creation attempt
		t.Logf("API correctly rejected cross-org station creation with 403")
	} else {
		t.Fatalf("expected 201 (forced to own org) or 403 (rejected), got %d: %v", resp.StatusCode, data)
	}
}
