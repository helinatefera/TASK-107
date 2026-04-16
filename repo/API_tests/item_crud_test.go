package API_tests

import (
	"net/http"
	"testing"
)

func TestCategoryAndItemCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "item_crud")

	// 1. Create a category
	catName := uniqueCode("cat")
	resp := doRequest(t, "POST", "/api/v1/categories", map[string]interface{}{
		"name": catName,
	}, adminToken)
	catData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create category, got %d: %v", resp.StatusCode, catData)
	}
	catID := catData["id"].(string)

	// 2. Create a unit
	unitName := uniqueCode("unit")
	unitSymbol := uniqueCode("sym")
	resp = doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   unitName,
		"symbol": unitSymbol,
	}, adminToken)
	unitData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create unit, got %d: %v", resp.StatusCode, unitData)
	}
	unitID := unitData["id"].(string)

	// 3. Create an item referencing the category and unit
	sku := uniqueCode("SKU")
	itemName := uniqueCode("item")
	resp = doRequest(t, "POST", "/api/v1/items", map[string]interface{}{
		"sku":          sku,
		"item_name":    itemName,
		"category_id":  catID,
		"base_unit_id": unitID,
		"description":  "Test item description",
	}, adminToken)
	itemData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create item, got %d: %v", resp.StatusCode, itemData)
	}
	itemID := itemData["id"].(string)

	// 4. List items and verify the created item is found
	resp = doRequest(t, "GET", "/api/v1/items?limit=100&offset=0", nil, adminToken)
	items := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list items, got %d", resp.StatusCode)
	}
	found := false
	for _, it := range items {
		m := it.(map[string]interface{})
		if m["id"] == itemID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created item not found in items list")
	}

	// 5. Get item by ID and verify fields
	resp = doRequest(t, "GET", "/api/v1/items/"+itemID, nil, adminToken)
	itemDetail := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on GET item, got %d: %v", resp.StatusCode, itemDetail)
	}
	if itemDetail["item_name"] != itemName {
		t.Fatalf("expected item_name %q, got %v", itemName, itemDetail["item_name"])
	}
	if itemDetail["sku"] != sku {
		t.Fatalf("expected sku %q, got %v", sku, itemDetail["sku"])
	}
	if itemDetail["description"] != "Test item description" {
		t.Fatalf("expected description 'Test item description', got %v", itemDetail["description"])
	}

	// 6. Update item name
	updatedName := uniqueCode("updated")
	resp = doRequest(t, "PUT", "/api/v1/items/"+itemID, map[string]interface{}{
		"item_name": updatedName,
	}, adminToken)
	updateData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on PUT item, got %d: %v", resp.StatusCode, updateData)
	}
	if updateData["item_name"] != updatedName {
		t.Fatalf("expected updated item_name %q, got %v", updatedName, updateData["item_name"])
	}

	// 7. Delete item
	resp = doRequest(t, "DELETE", "/api/v1/items/"+itemID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete item, got %d", resp.StatusCode)
	}

	// 8. Confirm item is gone (404)
	resp = doRequest(t, "GET", "/api/v1/items/"+itemID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 on deleted item, got %d", resp.StatusCode)
	}
}

func TestItemWithSubcategory(t *testing.T) {
	_, adminToken := createAdminUser(t, "item_subcat")

	// 1. Create parent category
	parentName := uniqueCode("parent")
	resp := doRequest(t, "POST", "/api/v1/categories", map[string]interface{}{
		"name": parentName,
	}, adminToken)
	parentData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create parent category, got %d: %v", resp.StatusCode, parentData)
	}
	parentID := parentData["id"].(string)

	// 2. Create child category with parent_id
	childName := uniqueCode("child")
	resp = doRequest(t, "POST", "/api/v1/categories", map[string]interface{}{
		"name":      childName,
		"parent_id": parentID,
	}, adminToken)
	childData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create child category, got %d: %v", resp.StatusCode, childData)
	}
	childID := childData["id"].(string)

	// 3. Create a unit for the item
	unitName := uniqueCode("unit")
	unitSymbol := uniqueCode("sym")
	resp = doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   unitName,
		"symbol": unitSymbol,
	}, adminToken)
	unitData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create unit, got %d: %v", resp.StatusCode, unitData)
	}
	unitID := unitData["id"].(string)

	// 4. Create item in child category
	sku := uniqueCode("SKU")
	itemName := uniqueCode("item")
	resp = doRequest(t, "POST", "/api/v1/items", map[string]interface{}{
		"sku":          sku,
		"item_name":    itemName,
		"category_id":  childID,
		"base_unit_id": unitID,
	}, adminToken)
	itemData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create item, got %d: %v", resp.StatusCode, itemData)
	}
	if itemData["category_id"] != childID {
		t.Fatalf("expected category_id %q, got %v", childID, itemData["category_id"])
	}

	// 5. List categories and verify both parent and child exist
	resp = doRequest(t, "GET", "/api/v1/categories", nil, adminToken)
	categories := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list categories, got %d", resp.StatusCode)
	}
	foundParent := false
	foundChild := false
	for _, c := range categories {
		m := c.(map[string]interface{})
		if m["id"] == parentID {
			foundParent = true
		}
		if m["id"] == childID {
			foundChild = true
		}
	}
	if !foundParent {
		t.Fatal("parent category not found in categories list")
	}
	if !foundChild {
		t.Fatal("child category not found in categories list")
	}
}

func TestUnitAndConversionCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "unit_conv")

	// 1. Create unit: kg
	resp := doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   uniqueCode("kg"),
		"symbol": uniqueCode("kg"),
	}, adminToken)
	kgData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create kg unit, got %d: %v", resp.StatusCode, kgData)
	}
	kgID := kgData["id"].(string)

	// 2. Create unit: lb
	resp = doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   uniqueCode("lb"),
		"symbol": uniqueCode("lb"),
	}, adminToken)
	lbData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create lb unit, got %d: %v", resp.StatusCode, lbData)
	}
	lbID := lbData["id"].(string)

	// 3. Create conversion: kg -> lb, factor 2.20462
	resp = doRequest(t, "POST", "/api/v1/units/conversions", map[string]interface{}{
		"from_unit_id": kgID,
		"to_unit_id":   lbID,
		"factor":       2.20462,
	}, adminToken)
	convData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create conversion, got %d: %v", resp.StatusCode, convData)
	}

	// 4. List units and verify both exist
	resp = doRequest(t, "GET", "/api/v1/units", nil, adminToken)
	units := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list units, got %d", resp.StatusCode)
	}
	foundKg := false
	foundLb := false
	for _, u := range units {
		m := u.(map[string]interface{})
		if m["id"] == kgID {
			foundKg = true
		}
		if m["id"] == lbID {
			foundLb = true
		}
	}
	if !foundKg {
		t.Fatal("kg unit not found in units list")
	}
	if !foundLb {
		t.Fatal("lb unit not found in units list")
	}

	// 5. List conversions and verify the conversion exists with correct factor
	resp = doRequest(t, "GET", "/api/v1/units/conversions", nil, adminToken)
	conversions := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list conversions, got %d", resp.StatusCode)
	}
	foundConv := false
	for _, c := range conversions {
		m := c.(map[string]interface{})
		if m["from_unit_id"] == kgID && m["to_unit_id"] == lbID {
			// decimal.Decimal serializes as string in JSON
			factorStr, ok := m["factor"].(string)
			if !ok {
				t.Fatalf("expected factor to be a string (decimal), got %T: %v", m["factor"], m["factor"])
			}
			if factorStr != "2.20462" {
				t.Fatalf("expected factor '2.20462', got %q", factorStr)
			}
			foundConv = true
			break
		}
	}
	if !foundConv {
		t.Fatal("conversion kg->lb not found in conversions list")
	}
}

func TestUserCanListItemsButNotCreate(t *testing.T) {
	_, adminToken := createAdminUser(t, "item_user_perm")

	// Create a unit so we have a valid base_unit_id for the POST attempt
	resp := doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   uniqueCode("unit"),
		"symbol": uniqueCode("sym"),
	}, adminToken)
	unitData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create unit, got %d: %v", resp.StatusCode, unitData)
	}
	unitID := unitData["id"].(string)

	// Register a regular user
	userEmail := uniqueEmail("reguser")
	userToken := registerAndLogin(t, userEmail)

	// Regular user can GET /items (200)
	resp = doRequest(t, "GET", "/api/v1/items", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on user GET /items, got %d", resp.StatusCode)
	}

	// Regular user can GET /categories (200)
	resp = doRequest(t, "GET", "/api/v1/categories", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on user GET /categories, got %d", resp.StatusCode)
	}

	// Regular user cannot POST /items (403)
	resp = doRequest(t, "POST", "/api/v1/items", map[string]interface{}{
		"sku":          uniqueCode("SKU"),
		"item_name":    uniqueCode("item"),
		"base_unit_id": unitID,
	}, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on user POST /items, got %d", resp.StatusCode)
	}
}

func TestMerchantCanCreateItem(t *testing.T) {
	_, adminToken := createAdminUser(t, "item_merchant")

	// Create org and merchant user
	orgID := adminCreateOrg(t, adminToken, "merch_org")
	_, merchantToken := createMerchantUser(t, "merch_item", orgID)

	// Admin creates a unit for use
	resp := doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   uniqueCode("unit"),
		"symbol": uniqueCode("sym"),
	}, adminToken)
	unitData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create unit, got %d: %v", resp.StatusCode, unitData)
	}
	unitID := unitData["id"].(string)

	// Merchant can POST /items (201)
	sku := uniqueCode("SKU")
	resp = doRequest(t, "POST", "/api/v1/items", map[string]interface{}{
		"sku":          sku,
		"item_name":    uniqueCode("item"),
		"base_unit_id": unitID,
	}, merchantToken)
	itemData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on merchant POST /items, got %d: %v", resp.StatusCode, itemData)
	}
	itemID := itemData["id"].(string)

	// Merchant cannot DELETE /items/:id (403, admin-only)
	resp = doRequest(t, "DELETE", "/api/v1/items/"+itemID, nil, merchantToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on merchant DELETE /items, got %d", resp.StatusCode)
	}
}

func TestCreateItemMissingRequiredFields(t *testing.T) {
	_, adminToken := createAdminUser(t, "item_valid")

	// Create a unit so we have a valid base_unit_id
	resp := doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   uniqueCode("unit"),
		"symbol": uniqueCode("sym"),
	}, adminToken)
	unitData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create unit, got %d: %v", resp.StatusCode, unitData)
	}
	unitID := unitData["id"].(string)

	// Missing sku -> 400
	resp = doRequest(t, "POST", "/api/v1/items", map[string]interface{}{
		"item_name":    uniqueCode("item"),
		"base_unit_id": unitID,
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 on missing sku, got %d", resp.StatusCode)
	}

	// Missing item_name -> 400
	resp = doRequest(t, "POST", "/api/v1/items", map[string]interface{}{
		"sku":          uniqueCode("SKU"),
		"base_unit_id": unitID,
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 on missing item_name, got %d", resp.StatusCode)
	}

	// Missing base_unit_id -> 400
	resp = doRequest(t, "POST", "/api/v1/items", map[string]interface{}{
		"sku":       uniqueCode("SKU"),
		"item_name": uniqueCode("item"),
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 on missing base_unit_id, got %d", resp.StatusCode)
	}
}

func TestUserCannotCreateCategory(t *testing.T) {
	// Register a regular user
	userEmail := uniqueEmail("catuser")
	userToken := registerAndLogin(t, userEmail)

	// Regular user cannot POST /categories (403)
	resp := doRequest(t, "POST", "/api/v1/categories", map[string]interface{}{
		"name": uniqueCode("cat"),
	}, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on user POST /categories, got %d", resp.StatusCode)
	}
}

func TestUserCannotCreateUnit(t *testing.T) {
	_, adminToken := createAdminUser(t, "unit_perm")

	// Register a regular user
	userEmail := uniqueEmail("unituser")
	userToken := registerAndLogin(t, userEmail)

	// Regular user cannot POST /units (403)
	resp := doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   uniqueCode("unit"),
		"symbol": uniqueCode("sym"),
	}, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on user POST /units, got %d", resp.StatusCode)
	}

	// Create org and merchant user
	orgID := adminCreateOrg(t, adminToken, "unit_merch_org")
	_, merchantToken := createMerchantUser(t, "unit_merch", orgID)

	// Merchant cannot POST /units (403)
	resp = doRequest(t, "POST", "/api/v1/units", map[string]interface{}{
		"name":   uniqueCode("unit"),
		"symbol": uniqueCode("sym"),
	}, merchantToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on merchant POST /units, got %d", resp.StatusCode)
	}
}
