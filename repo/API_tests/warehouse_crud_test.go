package API_tests

import (
	"net/http"
	"testing"
)

func TestAdminWarehouseCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "wh_crud")

	// 1. Create org
	orgID := adminCreateOrg(t, adminToken, "whorg")

	// 2. Create warehouse
	resp := doRequest(t, "POST", "/api/v1/warehouses", map[string]interface{}{
		"org_id":   orgID,
		"name":     "WH1",
		"timezone": "UTC",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create warehouse, got %d: %v", resp.StatusCode, data)
	}
	whID := data["id"].(string)

	// 3. GET warehouse
	resp = doRequest(t, "GET", "/api/v1/warehouses/"+whID, nil, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on GET warehouse, got %d: %v", resp.StatusCode, data)
	}
	if data["name"] != "WH1" {
		t.Fatalf("expected name WH1, got %v", data["name"])
	}

	// 4. PUT warehouse (update name)
	resp = doRequest(t, "PUT", "/api/v1/warehouses/"+whID, map[string]interface{}{
		"name": "WH1 Updated",
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on PUT warehouse, got %d: %v", resp.StatusCode, data)
	}
	if data["name"] != "WH1 Updated" {
		t.Fatalf("expected updated name, got %v", data["name"])
	}

	// 5. Create zone
	resp = doRequest(t, "POST", "/api/v1/warehouses/"+whID+"/zones", map[string]interface{}{
		"name": "Zone A",
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create zone, got %d: %v", resp.StatusCode, data)
	}
	zoneID := data["id"].(string)

	// 6. List zones
	resp = doRequest(t, "GET", "/api/v1/warehouses/"+whID+"/zones", nil, adminToken)
	zones := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list zones, got %d", resp.StatusCode)
	}
	if len(zones) < 1 {
		t.Fatal("expected at least 1 zone")
	}

	// 7. Create bin
	resp = doRequest(t, "POST", "/api/v1/zones/"+zoneID+"/bins", map[string]interface{}{
		"bin_code": "BIN-001",
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create bin, got %d: %v", resp.StatusCode, data)
	}

	// 8. List bins
	resp = doRequest(t, "GET", "/api/v1/zones/"+zoneID+"/bins", nil, adminToken)
	bins := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list bins, got %d", resp.StatusCode)
	}
	if len(bins) < 1 {
		t.Fatal("expected at least 1 bin")
	}

	// 9. Delete warehouse
	resp = doRequest(t, "DELETE", "/api/v1/warehouses/"+whID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete warehouse, got %d", resp.StatusCode)
	}
}
