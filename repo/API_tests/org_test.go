package API_tests

import (
	"net/http"
	"testing"
)

func TestCreateOrg_NoAuth(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]string{
		"name": "Test Org",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 401 {
		t.Fatalf("expected error code 401, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestCreateOrg_AsRegularUser(t *testing.T) {
	email := uniqueEmail("org_user")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]string{
		"name": "Test Org",
	}, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected error code 403, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestListOrgs_AsRegularUser(t *testing.T) {
	email := uniqueEmail("org_list_user")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "GET", "/api/v1/orgs", nil, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected error code 403, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestOrgErrorFormat_403(t *testing.T) {
	email := uniqueEmail("org_err_fmt")
	token := registerAndLogin(t, email)

	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]string{
		"name": "Test Org",
	}, token)
	data := parseJSON(t, resp)

	// Verify proper error format for 403
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}

	code, codeOk := data["code"].(float64)
	if !codeOk {
		t.Fatal("error response missing 'code' field")
	}
	if int(code) != 403 {
		t.Fatalf("expected code 403, got %v", code)
	}

	msg, msgOk := data["msg"].(string)
	if !msgOk {
		t.Fatal("error response missing 'msg' field")
	}
	if msg == "" {
		t.Fatal("error msg should not be empty")
	}
}
