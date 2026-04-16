package API_tests

import (
	"net/http"
	"strings"
	"testing"
)

func TestSupplierCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "supp_crud")
	orgID := adminCreateOrg(t, adminToken, "supp_org")

	// 1. Create supplier with all fields
	suppName := uniqueCode("Acme")
	resp := doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id":        orgID,
		"name":          suppName,
		"tax_id":        "TAX123456789",
		"contact_email": "acme@example.com",
		"address":       "123 Main St",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create supplier, got %d: %v", resp.StatusCode, data)
	}
	supplierID := data["id"].(string)

	// 2. List suppliers and verify found
	resp = doRequest(t, "GET", "/api/v1/suppliers?limit=100", nil, adminToken)
	suppliers := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list suppliers, got %d", resp.StatusCode)
	}
	found := false
	for _, s := range suppliers {
		m := s.(map[string]interface{})
		if m["id"] == supplierID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created supplier not found in list")
	}

	// 3. Get supplier by ID and verify all fields
	resp = doRequest(t, "GET", "/api/v1/suppliers/"+supplierID, nil, adminToken)
	detail := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on get supplier, got %d: %v", resp.StatusCode, detail)
	}
	if detail["name"] != suppName {
		t.Fatalf("expected name %q, got %v", suppName, detail["name"])
	}
	// Admin should see full tax_id
	if detail["tax_id"] != "TAX123456789" {
		t.Fatalf("expected admin to see full tax_id, got %v", detail["tax_id"])
	}

	// 4. Update supplier name and address
	updatedName := uniqueCode("AcmeUpdated")
	resp = doRequest(t, "PUT", "/api/v1/suppliers/"+supplierID, map[string]interface{}{
		"name":    updatedName,
		"address": "456 Oak Ave",
	}, adminToken)
	updateData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update supplier, got %d: %v", resp.StatusCode, updateData)
	}
	if updateData["name"] != updatedName {
		t.Fatalf("expected updated name %q, got %v", updatedName, updateData["name"])
	}
}

func TestCarrierCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "carr_crud")
	orgID := adminCreateOrg(t, adminToken, "carr_org")

	// 1. Create carrier
	carrName := uniqueCode("FastShip")
	resp := doRequest(t, "POST", "/api/v1/carriers", map[string]interface{}{
		"org_id":        orgID,
		"name":          carrName,
		"tax_id":        "CTAX987654321",
		"contact_email": "fast@ship.com",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create carrier, got %d: %v", resp.StatusCode, data)
	}
	carrierID := data["id"].(string)

	// 2. List carriers and verify found
	resp = doRequest(t, "GET", "/api/v1/carriers?limit=100", nil, adminToken)
	carriers := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list carriers, got %d", resp.StatusCode)
	}
	found := false
	for _, c := range carriers {
		m := c.(map[string]interface{})
		if m["id"] == carrierID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created carrier not found in list")
	}

	// 3. Get carrier by ID
	resp = doRequest(t, "GET", "/api/v1/carriers/"+carrierID, nil, adminToken)
	detail := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on get carrier, got %d: %v", resp.StatusCode, detail)
	}
	if detail["name"] != carrName {
		t.Fatalf("expected name %q, got %v", carrName, detail["name"])
	}

	// 4. Update carrier name
	updatedName := uniqueCode("FastShipV2")
	resp = doRequest(t, "PUT", "/api/v1/carriers/"+carrierID, map[string]interface{}{
		"name": updatedName,
	}, adminToken)
	updateData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update carrier, got %d: %v", resp.StatusCode, updateData)
	}
	if updateData["name"] != updatedName {
		t.Fatalf("expected updated name %q, got %v", updatedName, updateData["name"])
	}
}

func TestMerchantSupplierOrgScoping(t *testing.T) {
	_, adminToken := createAdminUser(t, "supp_scope")

	orgA := adminCreateOrg(t, adminToken, "supp_orgA")
	orgB := adminCreateOrg(t, adminToken, "supp_orgB")

	_, merchantToken := createMerchantUser(t, "supp_merch", orgA)

	// Merchant creates supplier (org_id overridden to their own org)
	resp := doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id": orgA,
		"name":   uniqueCode("MerchSupp"),
	}, merchantToken)
	merchSuppData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on merchant create supplier, got %d: %v", resp.StatusCode, merchSuppData)
	}

	// Admin creates supplier in org B
	resp = doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id": orgB,
		"name":   uniqueCode("OrgBSupp"),
	}, adminToken)
	orgBSuppData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on admin create supplier in orgB, got %d: %v", resp.StatusCode, orgBSuppData)
	}
	orgBSuppID := orgBSuppData["id"].(string)

	// Merchant lists suppliers — should only see org A supplier
	resp = doRequest(t, "GET", "/api/v1/suppliers?limit=100", nil, merchantToken)
	suppliers := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on merchant list suppliers, got %d", resp.StatusCode)
	}
	for _, s := range suppliers {
		m := s.(map[string]interface{})
		if m["id"] == orgBSuppID {
			t.Fatal("merchant should not see org B supplier in list")
		}
	}

	// Merchant tries GET on org B supplier → 403
	resp = doRequest(t, "GET", "/api/v1/suppliers/"+orgBSuppID, nil, merchantToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on merchant get org B supplier, got %d", resp.StatusCode)
	}
}

func TestSupplierPIIMasking(t *testing.T) {
	_, adminToken := createAdminUser(t, "supp_mask")
	orgID := adminCreateOrg(t, adminToken, "mask_org")

	// Admin creates supplier with tax_id
	resp := doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id":        orgID,
		"name":          uniqueCode("MaskCorp"),
		"tax_id":        "TAX123456789",
		"contact_email": "mask@example.com",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create supplier, got %d: %v", resp.StatusCode, data)
	}
	supplierID := data["id"].(string)

	// Admin sees full tax_id
	resp = doRequest(t, "GET", "/api/v1/suppliers/"+supplierID, nil, adminToken)
	adminView := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if adminView["tax_id"] != "TAX123456789" {
		t.Fatalf("expected admin to see full tax_id 'TAX123456789', got %v", adminView["tax_id"])
	}

	// Merchant sees masked tax_id
	_, merchantToken := createMerchantUser(t, "supp_mask_merch", orgID)
	resp = doRequest(t, "GET", "/api/v1/suppliers/"+supplierID, nil, merchantToken)
	merchantView := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	maskedTax, ok := merchantView["tax_id"].(string)
	if !ok {
		t.Fatalf("expected tax_id string, got %T", merchantView["tax_id"])
	}
	if maskedTax == "TAX123456789" {
		t.Fatal("expected merchant to see masked tax_id, but got full value")
	}
	if !strings.Contains(maskedTax, "****") && !strings.HasSuffix(maskedTax, "6789") {
		t.Fatalf("tax_id masking looks wrong: %q", maskedTax)
	}
}

func TestSupplierDuplicateDetection(t *testing.T) {
	_, adminToken := createAdminUser(t, "supp_dup")
	orgID := adminCreateOrg(t, adminToken, "dup_org")

	suppName := uniqueCode("DupCorp")
	taxID := uniqueCode("TAX")

	// Create first supplier
	resp := doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id": orgID,
		"name":   suppName,
		"tax_id": taxID,
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		data := parseJSON(t, resp)
		t.Fatalf("expected 201, got %d: %v", resp.StatusCode, data)
	}
	resp.Body.Close()

	// Create duplicate supplier with same name and tax_id in same org → 409
	resp = doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id": orgID,
		"name":   suppName,
		"tax_id": taxID,
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		data := parseJSON(t, resp)
		t.Fatalf("expected 409 on duplicate supplier, got %d: %v", resp.StatusCode, data)
	}
}

func TestCarrierDuplicateDetection(t *testing.T) {
	_, adminToken := createAdminUser(t, "carr_dup")
	orgID := adminCreateOrg(t, adminToken, "carr_dup_org")

	carrName := uniqueCode("DupCarrier")
	taxID := uniqueCode("CTAX")

	// Create first carrier
	resp := doRequest(t, "POST", "/api/v1/carriers", map[string]interface{}{
		"org_id": orgID,
		"name":   carrName,
		"tax_id": taxID,
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		data := parseJSON(t, resp)
		t.Fatalf("expected 201, got %d: %v", resp.StatusCode, data)
	}
	resp.Body.Close()

	// Create duplicate carrier → 409
	resp = doRequest(t, "POST", "/api/v1/carriers", map[string]interface{}{
		"org_id": orgID,
		"name":   carrName,
		"tax_id": taxID,
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		data := parseJSON(t, resp)
		t.Fatalf("expected 409 on duplicate carrier, got %d: %v", resp.StatusCode, data)
	}
}

func TestUserCannotAccessSuppliers(t *testing.T) {
	userToken := registerAndLogin(t, uniqueEmail("supp_nouser"))

	// Regular user cannot GET /suppliers (403)
	resp := doRequest(t, "GET", "/api/v1/suppliers", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on user GET /suppliers, got %d", resp.StatusCode)
	}

	// Regular user cannot POST /suppliers (403)
	resp = doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id": "00000000-0000-0000-0000-000000000001",
		"name":   "blocked",
	}, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on user POST /suppliers, got %d", resp.StatusCode)
	}
}

func TestSupplierInvalidEmail(t *testing.T) {
	_, adminToken := createAdminUser(t, "supp_email")
	orgID := adminCreateOrg(t, adminToken, "email_org")

	// Create supplier with invalid email → 400
	resp := doRequest(t, "POST", "/api/v1/suppliers", map[string]interface{}{
		"org_id":        orgID,
		"name":          uniqueCode("BadEmail"),
		"contact_email": "not-an-email",
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		data := parseJSON(t, resp)
		t.Fatalf("expected 400 on invalid email, got %d: %v", resp.StatusCode, data)
	}
}
