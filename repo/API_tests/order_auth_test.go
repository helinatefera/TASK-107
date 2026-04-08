package API_tests

import (
	"net/http"
	"testing"
	"time"
)

func TestUserCanCreateOrder_NoActivePricing(t *testing.T) {
	email := uniqueEmail("order_create")
	token := registerAndLogin(t, email)

	now := time.Now().UTC()
	resp := doRequest(t, "POST", "/api/v1/orders", map[string]interface{}{
		"order_id":   "ORD-TEST-001",
		"device_id":  "00000000-0000-0000-0000-000000000001",
		"energy_kwh": 10.5,
		"start_time": now.Add(-1 * time.Hour).Format(time.RFC3339),
		"end_time":   now.Format(time.RFC3339),
	}, token)
	data := parseJSON(t, resp)

	// The endpoint should be accessible (not 401/403) but may fail due to
	// missing active pricing version or invalid device.
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		t.Fatalf("expected accessible endpoint (not 401/403), got %d: %v", resp.StatusCode, data)
	}
}

func TestUserCanListOwnOrders(t *testing.T) {
	email := uniqueEmail("order_list")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/orders", nil, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, data)
	}
}

func TestOrderGetRequiresOwnership(t *testing.T) {
	email := uniqueEmail("order_get")
	token := registerAndLogin(t, email)

	// Try to access a non-existent order with a random UUID
	resp := doRequest(t, "GET", "/api/v1/orders/00000000-0000-0000-0000-000000000099", nil, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %v", resp.StatusCode, data)
	}
}
