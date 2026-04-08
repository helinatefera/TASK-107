package API_tests

import (
	"net/http"
	"testing"
)

func TestNotificationFlow(t *testing.T) {
	_, adminToken := createAdminUser(t, "notif_flow")

	// 1. Admin creates notification template
	tmplCode := uniqueCode("TMPL")
	resp := doRequest(t, "POST", "/api/v1/notifications/templates", map[string]interface{}{
		"code":       tmplCode,
		"title_tmpl": "Hello {{name}}",
		"body_tmpl":  "Test body",
	}, adminToken)
	tmplData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create notification template, got %d: %v", resp.StatusCode, tmplData)
	}
	templateID := tmplData["id"].(string)

	// 2. Admin lists templates -> verify our template exists
	resp = doRequest(t, "GET", "/api/v1/notifications/templates", nil, adminToken)
	templates := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list templates, got %d", resp.StatusCode)
	}
	found := false
	for _, tmpl := range templates {
		m := tmpl.(map[string]interface{})
		if m["id"] == templateID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find template %s in list", templateID)
	}

	// 3. Regular user lists subscriptions -> 200
	userEmail := uniqueEmail("notif_user")
	userToken := registerAndLogin(t, userEmail)
	resp = doRequest(t, "GET", "/api/v1/notifications/subscriptions", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list subscriptions, got %d", resp.StatusCode)
	}

	// 4. Regular user opts out of template
	resp = doRequest(t, "PUT", "/api/v1/notifications/subscriptions/"+templateID, map[string]interface{}{
		"opted_out": true,
	}, userToken)
	subData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on opt-out, got %d: %v", resp.StatusCode, subData)
	}

	// 5. Admin views delivery stats
	resp = doRequest(t, "GET", "/api/v1/notifications/stats", nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on delivery stats, got %d", resp.StatusCode)
	}
}
