package API_tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCarouselCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "carousel_crud")
	orgID := adminCreateOrg(t, adminToken, "carousel_org")

	startTime := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	// 1. Create carousel slot with all fields
	// target_role="administrator" so admin can see it in list
	resp := doRequest(t, "POST", "/api/v1/content/carousel", map[string]interface{}{
		"org_id":      orgID,
		"title":       "Summer Promo",
		"image_url":   "https://example.com/banner.png",
		"link_url":    "https://example.com/promo",
		"priority":    10,
		"target_role": "administrator",
		"start_time":  startTime,
		"end_time":    endTime,
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create carousel, got %d: %v", resp.StatusCode, data)
	}
	carouselID := data["id"].(string)

	// 2. List carousels and verify the created item exists
	resp = doRequest(t, "GET", "/api/v1/content/carousel", nil, adminToken)
	items := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list carousel, got %d", resp.StatusCode)
	}
	found := false
	for _, item := range items {
		m := item.(map[string]interface{})
		if m["id"] == carouselID {
			found = true
			if m["title"] != "Summer Promo" {
				t.Fatalf("expected title 'Summer Promo', got %v", m["title"])
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected to find carousel %s in list", carouselID)
	}

	// 3. Update title and priority
	resp = doRequest(t, "PUT", "/api/v1/content/carousel/"+carouselID, map[string]interface{}{
		"title":    "Winter Promo",
		"priority": 20,
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update carousel, got %d: %v", resp.StatusCode, data)
	}
	if data["title"] != "Winter Promo" {
		t.Fatalf("expected updated title 'Winter Promo', got %v", data["title"])
	}
	if int(data["priority"].(float64)) != 20 {
		t.Fatalf("expected updated priority 20, got %v", data["priority"])
	}

	// 4. Delete carousel
	resp = doRequest(t, "DELETE", "/api/v1/content/carousel/"+carouselID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete carousel, got %d", resp.StatusCode)
	}

	// 5. Confirm list no longer includes the deleted carousel
	resp = doRequest(t, "GET", "/api/v1/content/carousel", nil, adminToken)
	items = parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list carousel after delete, got %d", resp.StatusCode)
	}
	for _, item := range items {
		m := item.(map[string]interface{})
		if m["id"] == carouselID {
			t.Fatalf("carousel %s should have been deleted but still appears in list", carouselID)
		}
	}
}

func TestCampaignCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "campaign_crud")
	orgID := adminCreateOrg(t, adminToken, "campaign_org")

	startTime := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	// 1. Create campaign with JSON content
	resp := doRequest(t, "POST", "/api/v1/content/campaigns", map[string]interface{}{
		"org_id":      orgID,
		"name":        "Flash Sale",
		"content":     json.RawMessage(`{"heading":"Sale","discount":10}`),
		"priority":    5,
		"target_role": "administrator",
		"start_time":  startTime,
		"end_time":    endTime,
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create campaign, got %d: %v", resp.StatusCode, data)
	}
	campaignID := data["id"].(string)

	// 2. List campaigns and verify content JSON
	resp = doRequest(t, "GET", "/api/v1/content/campaigns", nil, adminToken)
	items := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list campaigns, got %d", resp.StatusCode)
	}
	found := false
	for _, item := range items {
		m := item.(map[string]interface{})
		if m["id"] == campaignID {
			found = true
			if m["name"] != "Flash Sale" {
				t.Fatalf("expected name 'Flash Sale', got %v", m["name"])
			}
			content, ok := m["content"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected content to be a JSON object, got %T", m["content"])
			}
			if content["heading"] != "Sale" {
				t.Fatalf("expected content heading 'Sale', got %v", content["heading"])
			}
			if int(content["discount"].(float64)) != 10 {
				t.Fatalf("expected content discount 10, got %v", content["discount"])
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected to find campaign %s in list", campaignID)
	}

	// 3. Update campaign name
	resp = doRequest(t, "PUT", "/api/v1/content/campaigns/"+campaignID, map[string]interface{}{
		"name": "Mega Sale",
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update campaign, got %d: %v", resp.StatusCode, data)
	}
	if data["name"] != "Mega Sale" {
		t.Fatalf("expected updated name 'Mega Sale', got %v", data["name"])
	}

	// 4. Delete campaign
	resp = doRequest(t, "DELETE", "/api/v1/content/campaigns/"+campaignID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete campaign, got %d", resp.StatusCode)
	}
}

func TestRankingCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "ranking_crud")
	orgID := adminCreateOrg(t, adminToken, "ranking_org")

	startTime := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	// 1. Create ranking
	resp := doRequest(t, "POST", "/api/v1/content/rankings", map[string]interface{}{
		"org_id":      orgID,
		"entity_type": "station",
		"entity_id":   "00000000-0000-0000-0000-000000000001",
		"score":       100,
		"priority":    1,
		"target_role": "administrator",
		"start_time":  startTime,
		"end_time":    endTime,
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create ranking, got %d: %v", resp.StatusCode, data)
	}
	rankingID := data["id"].(string)

	// 2. List rankings and verify
	resp = doRequest(t, "GET", "/api/v1/content/rankings", nil, adminToken)
	items := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list rankings, got %d", resp.StatusCode)
	}
	found := false
	for _, item := range items {
		m := item.(map[string]interface{})
		if m["id"] == rankingID {
			found = true
			if m["entity_type"] != "station" {
				t.Fatalf("expected entity_type 'station', got %v", m["entity_type"])
			}
			if int(m["score"].(float64)) != 100 {
				t.Fatalf("expected score 100, got %v", m["score"])
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected to find ranking %s in list", rankingID)
	}

	// 3. Update score to 200
	resp = doRequest(t, "PUT", "/api/v1/content/rankings/"+rankingID, map[string]interface{}{
		"score": 200,
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update ranking, got %d: %v", resp.StatusCode, data)
	}
	if int(data["score"].(float64)) != 200 {
		t.Fatalf("expected updated score 200, got %v", data["score"])
	}

	// 4. Delete ranking
	resp = doRequest(t, "DELETE", "/api/v1/content/rankings/"+rankingID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete ranking, got %d", resp.StatusCode)
	}
}

func TestMerchantCanCreateButNotDeleteCarousel(t *testing.T) {
	_, adminToken := createAdminUser(t, "merch_carousel")
	orgID := adminCreateOrg(t, adminToken, "merch_car_org")
	_, merchantToken := createMerchantUser(t, "merch_car", orgID)

	startTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	// Merchant creates carousel (org_id forced to their org)
	resp := doRequest(t, "POST", "/api/v1/content/carousel", map[string]interface{}{
		"org_id":     orgID,
		"title":      "Merchant Promo",
		"priority":   1,
		"start_time": startTime,
		"end_time":   endTime,
	}, merchantToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on merchant create carousel, got %d: %v", resp.StatusCode, data)
	}
	carouselID := data["id"].(string)

	// Merchant tries to delete -> 403
	resp = doRequest(t, "DELETE", "/api/v1/content/carousel/"+carouselID, nil, merchantToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on merchant delete carousel, got %d", resp.StatusCode)
	}
}

func TestMerchantCannotAccessOtherOrgContent(t *testing.T) {
	_, adminToken := createAdminUser(t, "crossorg_content")

	orgA := adminCreateOrg(t, adminToken, "content_orgA")
	orgB := adminCreateOrg(t, adminToken, "content_orgB")

	// Create merchant in org A
	_, merchantToken := createMerchantUser(t, "crossorg_cont_merch", orgA)

	startTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	// Admin creates carousel in org B
	resp := doRequest(t, "POST", "/api/v1/content/carousel", map[string]interface{}{
		"org_id":     orgB,
		"title":      "OrgB Carousel",
		"priority":   1,
		"start_time": startTime,
		"end_time":   endTime,
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on admin create carousel in orgB, got %d: %v", resp.StatusCode, data)
	}
	carouselID := data["id"].(string)

	// Merchant in org A tries to PUT on org B carousel -> 403
	resp = doRequest(t, "PUT", "/api/v1/content/carousel/"+carouselID, map[string]interface{}{
		"title": "Hijacked",
	}, merchantToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		data = parseJSON(t, resp)
		t.Fatalf("expected 403 for cross-org carousel update, got %d: %v", resp.StatusCode, data)
	}
}

func TestUserCanListContent(t *testing.T) {
	userEmail := uniqueEmail("content_user")
	userToken := registerAndLogin(t, userEmail)

	// Regular user can list carousel
	resp := doRequest(t, "GET", "/api/v1/content/carousel", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on user list carousel, got %d", resp.StatusCode)
	}

	// Regular user can list campaigns
	resp = doRequest(t, "GET", "/api/v1/content/campaigns", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on user list campaigns, got %d", resp.StatusCode)
	}

	// Regular user can list rankings
	resp = doRequest(t, "GET", "/api/v1/content/rankings", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on user list rankings, got %d", resp.StatusCode)
	}
}

func TestInvalidTargetRole(t *testing.T) {
	_, adminToken := createAdminUser(t, "invalid_role")
	orgID := adminCreateOrg(t, adminToken, "invalid_role_org")

	startTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	// Create carousel with invalid target_role
	resp := doRequest(t, "POST", "/api/v1/content/carousel", map[string]interface{}{
		"org_id":      orgID,
		"title":       "Bad Role Carousel",
		"priority":    1,
		"target_role": "superadmin",
		"start_time":  startTime,
		"end_time":    endTime,
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		data := parseJSON(t, resp)
		t.Fatalf("expected 400 for invalid target_role, got %d: %v", resp.StatusCode, data)
	}
}

func TestCarouselMissingRequiredFields(t *testing.T) {
	_, adminToken := createAdminUser(t, "carousel_missing")
	orgID := adminCreateOrg(t, adminToken, "carousel_miss_org")

	startTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	// Missing title
	resp := doRequest(t, "POST", "/api/v1/content/carousel", map[string]interface{}{
		"org_id":     orgID,
		"priority":   1,
		"start_time": startTime,
		"end_time":   endTime,
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		data := parseJSON(t, resp)
		t.Fatalf("expected 400 for missing title, got %d: %v", resp.StatusCode, data)
	}

	// Missing start_time
	resp = doRequest(t, "POST", "/api/v1/content/carousel", map[string]interface{}{
		"org_id":   orgID,
		"title":    "No Start Time",
		"priority": 1,
		"end_time": endTime,
	}, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		data := parseJSON(t, resp)
		t.Fatalf("expected 400 for missing start_time, got %d: %v", resp.StatusCode, data)
	}
}
