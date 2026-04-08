package API_tests

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestProtectedEndpoint_NoToken(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/auth/logout", nil, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 401 {
		t.Fatalf("expected error code 401, got %v", data["code"])
	}
}

func TestProtectedEndpoint_InvalidToken(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/auth/logout", nil, "invalid-token-abc123")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if code, ok := data["code"].(float64); !ok || int(code) != 401 {
		t.Fatalf("expected error code 401, got %v", data["code"])
	}
}

func TestErrorResponses_NoStackTraces(t *testing.T) {
	// Send various bad requests and verify no stack traces leak
	endpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
		token  string
	}{
		{
			name:   "bad register",
			method: "POST",
			path:   "/api/v1/auth/register",
			body:   map[string]string{"email": "bad"},
		},
		{
			name:   "bad login",
			method: "POST",
			path:   "/api/v1/auth/login",
			body:   map[string]string{"email": "nobody@test.com", "password": "x", "device_id": "d"},
		},
		{
			name:   "no auth on protected",
			method: "POST",
			path:   "/api/v1/auth/logout",
		},
		{
			name:   "invalid token",
			method: "POST",
			path:   "/api/v1/auth/refresh",
			token:  "bogus-token",
		},
	}

	dangerousStrings := []string{"goroutine", "panic", ".go:"}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			resp := doRequest(t, ep.method, ep.path, ep.body, ep.token)
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			bodyStr := string(bodyBytes)
			for _, dangerous := range dangerousStrings {
				if strings.Contains(bodyStr, dangerous) {
					t.Errorf("response body contains %q which may indicate a stack trace leak: %s", dangerous, bodyStr)
				}
			}
		})
	}
}

func TestAllErrorResponses_HaveCodeAndMsg(t *testing.T) {
	endpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
		token  string
	}{
		{
			name:   "register empty body",
			method: "POST",
			path:   "/api/v1/auth/register",
			body:   map[string]string{},
		},
		{
			name:   "login empty body",
			method: "POST",
			path:   "/api/v1/auth/login",
			body:   map[string]string{},
		},
		{
			name:   "logout no token",
			method: "POST",
			path:   "/api/v1/auth/logout",
		},
		{
			name:   "refresh invalid token",
			method: "POST",
			path:   "/api/v1/auth/refresh",
			token:  "invalid-token",
		},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			resp := doRequest(t, ep.method, ep.path, ep.body, ep.token)
			data := parseJSON(t, resp)

			if _, ok := data["code"]; !ok {
				t.Errorf("error response missing 'code' field, got: %v", data)
			}
			if _, ok := data["msg"]; !ok {
				t.Errorf("error response missing 'msg' field, got: %v", data)
			}
		})
	}
}
