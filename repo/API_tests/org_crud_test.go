package API_tests

import (
	"net/http"
	"testing"
)

// TestOrgCRUDLifecycle covers the success paths for create, get, list, update, delete.
func TestOrgCRUDLifecycle(t *testing.T) {
	_, adminToken := createAdminUser(t, "org_crud")

	orgCode := uniqueCode("ORG")
	orgName := uniqueCode("TestOrg")

	// 1. Create org
	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code": orgCode,
		"name":     orgName,
		"timezone": "America/New_York",
		"tax_id":   "ORG-TAX-001",
		"address":  "100 Corp Blvd",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create org, got %d: %v", resp.StatusCode, data)
	}
	orgID := data["id"].(string)
	if data["org_code"] != orgCode {
		t.Fatalf("expected org_code %q, got %v", orgCode, data["org_code"])
	}
	if data["name"] != orgName {
		t.Fatalf("expected name %q, got %v", orgName, data["name"])
	}
	if data["timezone"] != "America/New_York" {
		t.Fatalf("expected timezone America/New_York, got %v", data["timezone"])
	}

	// 2. Get org by ID — verify fields
	resp = doRequest(t, "GET", "/api/v1/orgs/"+orgID, nil, adminToken)
	detail := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on get org, got %d: %v", resp.StatusCode, detail)
	}
	if detail["id"] != orgID {
		t.Fatalf("expected id %q, got %v", orgID, detail["id"])
	}
	if detail["name"] != orgName {
		t.Fatalf("expected name %q, got %v", orgName, detail["name"])
	}
	if _, ok := detail["created_at"].(string); !ok {
		t.Fatal("expected created_at field")
	}
	if _, ok := detail["updated_at"].(string); !ok {
		t.Fatal("expected updated_at field")
	}

	// 3. List orgs — verify created org appears
	resp = doRequest(t, "GET", "/api/v1/orgs?limit=100", nil, adminToken)
	orgs := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list orgs, got %d", resp.StatusCode)
	}
	found := false
	for _, o := range orgs {
		m := o.(map[string]interface{})
		if m["id"] == orgID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created org not found in org list")
	}

	// 4. Update org name and timezone
	updatedName := uniqueCode("UpdatedOrg")
	resp = doRequest(t, "PUT", "/api/v1/orgs/"+orgID, map[string]interface{}{
		"name":     updatedName,
		"timezone": "Europe/London",
	}, adminToken)
	updateData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update org, got %d: %v", resp.StatusCode, updateData)
	}
	if updateData["name"] != updatedName {
		t.Fatalf("expected updated name %q, got %v", updatedName, updateData["name"])
	}
	if updateData["timezone"] != "Europe/London" {
		t.Fatalf("expected updated timezone Europe/London, got %v", updateData["timezone"])
	}

	// 5. Delete org
	resp = doRequest(t, "DELETE", "/api/v1/orgs/"+orgID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete org, got %d", resp.StatusCode)
	}

	// 6. Confirm deleted — GET returns 404
	resp = doRequest(t, "GET", "/api/v1/orgs/"+orgID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 on deleted org, got %d", resp.StatusCode)
	}
}

// TestOrgDuplicateOrgCode verifies creating two orgs with the same org_code returns 409.
func TestOrgDuplicateOrgCode(t *testing.T) {
	_, adminToken := createAdminUser(t, "org_dup")

	code := uniqueCode("DUPORG")

	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code": code,
		"name":     "First Org",
		"timezone": "UTC",
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		data := parseJSON(t, resp)
		t.Fatalf("expected 201, got %d: %v", resp.StatusCode, data)
	}
	resp.Body.Close()

	// Duplicate org_code → 409
	resp = doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code": code,
		"name":     "Second Org",
		"timezone": "UTC",
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		data := parseJSON(t, resp)
		t.Fatalf("expected 409 on duplicate org_code, got %d: %v", resp.StatusCode, data)
	}
}

// TestMerchantCanGetOwnOrg verifies a merchant can GET their assigned org.
func TestMerchantCanGetOwnOrg(t *testing.T) {
	_, adminToken := createAdminUser(t, "org_merch_get")
	orgID := adminCreateOrg(t, adminToken, "merch_own_org")
	_, merchantToken := createMerchantUser(t, "org_merch", orgID)

	resp := doRequest(t, "GET", "/api/v1/orgs/"+orgID, nil, merchantToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on merchant get own org, got %d: %v", resp.StatusCode, data)
	}
	if data["id"] != orgID {
		t.Fatalf("expected org id %q, got %v", orgID, data["id"])
	}
}

// TestMerchantCannotDeleteOrg verifies merchants cannot delete orgs (admin-only).
func TestMerchantCannotDeleteOrg(t *testing.T) {
	_, adminToken := createAdminUser(t, "org_merch_del")
	orgID := adminCreateOrg(t, adminToken, "merch_del_org")
	_, merchantToken := createMerchantUser(t, "org_del_merch", orgID)

	resp := doRequest(t, "DELETE", "/api/v1/orgs/"+orgID, nil, merchantToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on merchant delete org, got %d", resp.StatusCode)
	}
}
