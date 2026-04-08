package API_tests

import (
	"net/http"
	"strings"
	"testing"
)

func TestRegister_InvalidEmailFormat(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        "not-an-email",
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected error code 400, got %v", data["code"])
	}
	if _, ok := data["msg"].(string); !ok {
		t.Fatal("expected msg field in error response")
	}
}

func TestMalformedJSON(t *testing.T) {
	endpoints := []struct {
		name string
		path string
	}{
		{"register", "/api/v1/auth/register"},
		{"login", "/api/v1/auth/login"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", baseURL()+ep.path, strings.NewReader("{bad json"))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			data := parseJSON(t, resp)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}
			if code, ok := data["code"].(float64); !ok || int(code) != 400 {
				t.Fatalf("expected error code 400, got %v", data["code"])
			}
			if _, ok := data["msg"].(string); !ok {
				t.Fatal("expected msg field in error response")
			}
		})
	}
}

func TestEmptyBody(t *testing.T) {
	endpoints := []struct {
		name string
		path string
	}{
		{"register", "/api/v1/auth/register"},
		{"login", "/api/v1/auth/login"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", baseURL()+ep.path, strings.NewReader(""))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			data := parseJSON(t, resp)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}
			if code, ok := data["code"].(float64); !ok || int(code) != 400 {
				t.Fatalf("expected error code 400, got %v", data["code"])
			}
			if _, ok := data["msg"].(string); !ok {
				t.Fatal("expected msg field in error response")
			}
		})
	}
}

func TestValidationErrors_ConsistentFormat(t *testing.T) {
	// Test a variety of validation failures to ensure consistent error format
	tests := []struct {
		name string
		path string
		body interface{}
	}{
		{
			name: "register no email",
			path: "/api/v1/auth/register",
			body: map[string]string{
				"password":     "SecurePass123!",
				"display_name": "Test",
			},
		},
		{
			name: "register no password",
			path: "/api/v1/auth/register",
			body: map[string]string{
				"email":        uniqueEmail("val_nopw"),
				"display_name": "Test",
			},
		},
		{
			name: "register no display_name",
			path: "/api/v1/auth/register",
			body: map[string]string{
				"email":    uniqueEmail("val_nodn"),
				"password": "SecurePass123!",
			},
		},
		{
			name: "login no email",
			path: "/api/v1/auth/login",
			body: map[string]string{
				"password":  "SecurePass123!",
				"device_id": "test-device-001",
			},
		},
		{
			name: "login no password",
			path: "/api/v1/auth/login",
			body: map[string]string{
				"email":     uniqueEmail("val_nopwl"),
				"device_id": "test-device-001",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, "POST", tc.path, tc.body, "")
			data := parseJSON(t, resp)

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}

			code, codeOk := data["code"].(float64)
			if !codeOk {
				t.Errorf("missing 'code' field in response: %v", data)
			} else if int(code) != 400 {
				t.Errorf("expected code 400, got %v", code)
			}

			msg, msgOk := data["msg"].(string)
			if !msgOk {
				t.Errorf("missing 'msg' field in response: %v", data)
			} else if msg == "" {
				t.Error("msg should not be empty")
			}
		})
	}
}
