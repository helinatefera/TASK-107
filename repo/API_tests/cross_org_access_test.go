package API_tests

import (
	"net/http"
	"testing"
)

// TestMerchantCannotAccessOtherOrgStation verifies that a merchant from org A
// cannot access a station belonging to org B.
func TestMerchantCannotAccessOtherOrgStation(t *testing.T) {
	_, adminToken := createAdminUser(t, "crossorg_st")

	orgA := adminCreateOrg(t, adminToken, "orgA_st")
	orgB := adminCreateOrg(t, adminToken, "orgB_st")

	// Create station in org B
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgB, "name": "OrgB Station", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	// Merchant in org A tries to GET the station
	_, merchantToken := createMerchantUser(t, "crossorg_st_merch", orgA)
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, merchantToken)
	if resp.StatusCode != http.StatusForbidden {
		data := parseJSON(t, resp)
		t.Fatalf("expected 403 for cross-org station access, got %d: %v", resp.StatusCode, data)
	}
}

// TestMerchantCannotAccessOtherOrgWarehouse verifies cross-org warehouse isolation.
func TestMerchantCannotAccessOtherOrgWarehouse(t *testing.T) {
	_, adminToken := createAdminUser(t, "crossorg_wh")

	orgA := adminCreateOrg(t, adminToken, "orgA_wh")
	orgB := adminCreateOrg(t, adminToken, "orgB_wh")

	// Create warehouse in org B
	resp := doRequest(t, "POST", "/api/v1/warehouses", map[string]interface{}{
		"org_id": orgB, "name": "OrgB Warehouse", "timezone": "UTC",
	}, adminToken)
	whData := parseJSON(t, resp)
	whID := whData["id"].(string)

	// Merchant in org A tries to GET the warehouse
	_, merchantToken := createMerchantUser(t, "crossorg_wh_merch", orgA)
	resp = doRequest(t, "GET", "/api/v1/warehouses/"+whID, nil, merchantToken)
	if resp.StatusCode != http.StatusForbidden {
		data := parseJSON(t, resp)
		t.Fatalf("expected 403 for cross-org warehouse access, got %d: %v", resp.StatusCode, data)
	}
}

// TestMerchantCannotAccessOtherOrgPricingTemplate verifies cross-org pricing isolation.
func TestMerchantCannotAccessOtherOrgPricingTemplate(t *testing.T) {
	_, adminToken := createAdminUser(t, "crossorg_pr")

	orgA := adminCreateOrg(t, adminToken, "orgA_pr")
	orgB := adminCreateOrg(t, adminToken, "orgB_pr")

	// Create station + pricing template in org B
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgB, "name": "OrgB PricingStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	resp = doRequest(t, "POST", "/api/v1/pricing/templates", map[string]interface{}{
		"name": "OrgB Pricing", "station_id": stationID,
	}, adminToken)
	tmplData := parseJSON(t, resp)
	tmplID := tmplData["id"].(string)

	// Merchant in org A tries to GET the template
	_, merchantToken := createMerchantUser(t, "crossorg_pr_merch", orgA)
	resp = doRequest(t, "GET", "/api/v1/pricing/templates/"+tmplID, nil, merchantToken)
	if resp.StatusCode != http.StatusForbidden {
		data := parseJSON(t, resp)
		t.Fatalf("expected 403 for cross-org pricing template access, got %d: %v", resp.StatusCode, data)
	}
}

// TestParentOrgCanAccessChildOrgStation verifies that hierarchy-aware
// authorization allows a parent org merchant to access child org resources.
func TestParentOrgCanAccessChildOrgStation(t *testing.T) {
	_, adminToken := createAdminUser(t, "parentchild_st")

	parentOrg := adminCreateOrg(t, adminToken, "parent_st")

	// Create child org under parent
	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code":  uniqueCode("child_st"),
		"name":      "Child Org",
		"timezone":  "UTC",
		"parent_id": parentOrg,
	}, adminToken)
	childData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create child org: %d %v", resp.StatusCode, childData)
	}
	childOrg := childData["id"].(string)

	// Create station in child org
	resp = doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": childOrg, "name": "ChildStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	// Merchant in parent org should be able to access child org's station
	_, parentMerchantToken := createMerchantUser(t, "parent_st_merch", parentOrg)
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, parentMerchantToken)
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("parent org merchant should access child org station, got %d: %v", resp.StatusCode, data)
	}
}

// TestChildOrgCannotAccessParentOrgStation verifies that a child org merchant
// cannot access parent org resources (hierarchy is one-directional).
func TestChildOrgCannotAccessParentOrgStation(t *testing.T) {
	_, adminToken := createAdminUser(t, "childparent_st")

	parentOrg := adminCreateOrg(t, adminToken, "parent2_st")

	// Create child org
	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code":  uniqueCode("child2_st"),
		"name":      "Child Org 2",
		"timezone":  "UTC",
		"parent_id": parentOrg,
	}, adminToken)
	childData := parseJSON(t, resp)
	childOrg := childData["id"].(string)

	// Create station in parent org
	resp = doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": parentOrg, "name": "ParentStation", "timezone": "UTC",
	}, adminToken)
	stData := parseJSON(t, resp)
	stationID := stData["id"].(string)

	// Merchant in child org should NOT access parent org's station
	_, childMerchantToken := createMerchantUser(t, "child_st_merch", childOrg)
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, childMerchantToken)
	if resp.StatusCode != http.StatusForbidden {
		data := parseJSON(t, resp)
		t.Fatalf("child org merchant should not access parent org station, got %d: %v", resp.StatusCode, data)
	}
}

// TestMerchantListOnlyShowsOwnOrgResources verifies that list endpoints
// filter by org and don't leak cross-org data.
func TestMerchantListOnlyShowsOwnOrgResources(t *testing.T) {
	_, adminToken := createAdminUser(t, "list_iso")

	orgA := adminCreateOrg(t, adminToken, "listA")
	orgB := adminCreateOrg(t, adminToken, "listB")

	// Create stations in both orgs
	doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgA, "name": "StationA", "timezone": "UTC",
	}, adminToken)
	doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id": orgB, "name": "StationB", "timezone": "UTC",
	}, adminToken)

	// Merchant in org A lists stations — should only see org A's station
	_, merchantAToken := createMerchantUser(t, "list_iso_merch", orgA)
	resp := doRequest(t, "GET", "/api/v1/stations", nil, merchantAToken)
	stations := parseJSONArray(t, resp)

	for _, s := range stations {
		st := s.(map[string]interface{})
		if st["name"] == "StationB" {
			t.Fatal("merchant in org A should not see org B's station in list")
		}
	}
}
