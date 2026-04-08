package API_tests

import (
	"net/http"
	"testing"
)

func TestUserCanListOwnInbox(t *testing.T) {
	email := uniqueEmail("notif_inbox")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/notifications/inbox", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, data)
	}
}

func TestUserCanListSubscriptions(t *testing.T) {
	email := uniqueEmail("notif_subs")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/notifications/subscriptions", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, data)
	}
}

func TestMarkReadRequiresOwnership(t *testing.T) {
	email := uniqueEmail("notif_read")
	token := registerAndLogin(t, email)

	// Try to mark a non-existent notification as read
	resp := doRequest(t, "POST", "/api/v1/notifications/inbox/00000000-0000-0000-0000-000000000099/read", nil, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %v", resp.StatusCode, data)
	}
}

func TestMarkDismissRequiresOwnership(t *testing.T) {
	email := uniqueEmail("notif_dismiss")
	token := registerAndLogin(t, email)

	// Try to dismiss a non-existent notification
	resp := doRequest(t, "POST", "/api/v1/notifications/inbox/00000000-0000-0000-0000-000000000099/dismiss", nil, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %v", resp.StatusCode, data)
	}
}

func TestNonAdminCannotCreateTemplate(t *testing.T) {
	email := uniqueEmail("notif_tmpl")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/notifications/templates", map[string]string{
		"code":       "test-template",
		"title_tmpl": "Test Subject",
		"body_tmpl":  "Test Body",
	}, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %v", resp.StatusCode, data)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected error code 403, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestNonAdminCannotViewStats(t *testing.T) {
	email := uniqueEmail("notif_stats")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/notifications/stats", nil, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %v", resp.StatusCode, data)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected error code 403, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}
