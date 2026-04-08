package API_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func baseURL() string {
	if u := os.Getenv("API_BASE_URL"); u != "" {
		return u
	}
	return "http://localhost:8080"
}

func doRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL()+path, bodyReader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Device-Id", "test-device-001")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func parseJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(b, &result)
	return result
}

func registerAndLogin(t *testing.T, email string) string {
	t.Helper()
	// Register
	doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":        email,
		"password":     "SecurePass123!",
		"display_name": "Test User",
	}, "")
	// Login
	resp := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":     email,
		"password":  "SecurePass123!",
		"device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if token, ok := data["token"].(string); ok {
		return token
	}
	t.Fatal("failed to login")
	return ""
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@test.com", prefix, time.Now().UnixNano())
}

func parseJSONArray(t *testing.T, resp *http.Response) []interface{} {
	t.Helper()
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var result []interface{}
	json.Unmarshal(b, &result)
	return result
}

func promoteUser(t *testing.T, userID, role, adminToken string) {
	t.Helper()
	resp := doRequest(t, "PUT", "/api/v1/users/"+userID+"/role", map[string]string{
		"role": role,
	}, adminToken)
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("promoteUser failed: %d %v", resp.StatusCode, data)
	}
	resp.Body.Close()
}

func setUserOrg(t *testing.T, userID, orgID, adminToken string) {
	t.Helper()
	resp := doRequest(t, "PUT", "/api/v1/users/"+userID+"/org", map[string]string{
		"org_id": orgID,
	}, adminToken)
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("setUserOrg failed: %d %v", resp.StatusCode, data)
	}
	resp.Body.Close()
}

// registerGetIDAndLogin registers a user, returns (userID, token)
func registerGetIDAndLogin(t *testing.T, email string) (string, string) {
	t.Helper()
	resp := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email": email, "password": "SecurePass123!", "display_name": "Test User",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %d %v", resp.StatusCode, data)
	}
	userID := data["id"].(string)

	resp2 := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": "SecurePass123!", "device_id": "test-device-001",
	}, "")
	data2 := parseJSON(t, resp2)
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("login failed: %d %v", resp2.StatusCode, data2)
	}
	token := data2["token"].(string)
	return userID, token
}

// bootstrapAdmin creates the very first admin user via direct DB promotion (no admin token exists yet).
func bootstrapAdmin(t *testing.T, prefix string) (string, string) {
	t.Helper()
	email := uniqueEmail(prefix)
	userID, _ := registerGetIDAndLogin(t, email)
	// Direct DB for bootstrap only — no admin token exists yet
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://chargeops:chargeops@localhost:5433/chargeops?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec("UPDATE users SET role = 'administrator' WHERE id = $1", userID)
	if err != nil {
		t.Fatal(err)
	}
	// Re-login so the session reflects the new role
	resp := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": "SecurePass123!", "device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin re-login failed: %d %v", resp.StatusCode, data)
	}
	token := data["token"].(string)
	return userID, token
}

// createAdminUser registers a user, promotes to admin via API, returns (userID, token)
func createAdminUser(t *testing.T, prefix string) (string, string) {
	t.Helper()
	// Bootstrap a temporary admin to use the API
	_, bootstrapToken := bootstrapAdmin(t, prefix+"_boot")

	email := uniqueEmail(prefix)
	userID, _ := registerGetIDAndLogin(t, email)
	promoteUser(t, userID, "administrator", bootstrapToken)
	// Re-login so the session reflects the new role
	resp := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": "SecurePass123!", "device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin re-login failed: %d %v", resp.StatusCode, data)
	}
	token := data["token"].(string)
	return userID, token
}

// createMerchantUser registers, promotes to merchant, assigns to org via API, returns (userID, token)
func createMerchantUser(t *testing.T, prefix, orgID string) (string, string) {
	t.Helper()
	// Bootstrap a temporary admin to use the API
	_, bootstrapToken := bootstrapAdmin(t, prefix+"_boot")

	email := uniqueEmail(prefix)
	userID, _ := registerGetIDAndLogin(t, email)
	promoteUser(t, userID, "merchant", bootstrapToken)
	setUserOrg(t, userID, orgID, bootstrapToken)
	resp := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": "SecurePass123!", "device_id": "test-device-001",
	}, "")
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("merchant re-login failed: %d %v", resp.StatusCode, data)
	}
	token := data["token"].(string)
	return userID, token
}

// uniqueCode generates a unique code for orgs, devices, etc.
func uniqueCode(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// adminCreateOrg is a convenience that creates an org as admin, returns orgID
func adminCreateOrg(t *testing.T, token, prefix string) string {
	t.Helper()
	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code": uniqueCode(prefix),
		"name":     prefix + " Org",
		"timezone": "UTC",
	}, token)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create org failed: %d %v", resp.StatusCode, data)
	}
	return data["id"].(string)
}
