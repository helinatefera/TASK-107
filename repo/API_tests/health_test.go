package API_tests

import (
	"net/http"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	resp := doRequest(t, "GET", "/health", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	data := parseJSON(t, resp)
	status, ok := data["status"].(string)
	if !ok || status != "ok" {
		t.Fatalf("expected status=ok, got %v", data["status"])
	}
}
