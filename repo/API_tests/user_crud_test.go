package API_tests

import (
	"net/http"
	"testing"
)

// TestAdminListUsers verifies GET /users returns a list with expected fields.
func TestAdminListUsers(t *testing.T) {
	adminID, adminToken := createAdminUser(t, "user_list")

	resp := doRequest(t, "GET", "/api/v1/users?limit=10", nil, adminToken)
	users := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(users) == 0 {
		t.Fatal("expected at least 1 user in list")
	}

	// Verify the admin user appears in the list with expected fields
	found := false
	for _, u := range users {
		m := u.(map[string]interface{})
		if m["id"] == adminID {
			found = true
			if _, ok := m["email"].(string); !ok {
				t.Fatal("expected email field in user object")
			}
			if _, ok := m["role"].(string); !ok {
				t.Fatal("expected role field in user object")
			}
			if _, ok := m["created_at"].(string); !ok {
				t.Fatal("expected created_at field in user object")
			}
			break
		}
	}
	if !found {
		t.Fatal("admin user not found in user list")
	}
}

// TestNonAdminCannotListUsers verifies regular users are blocked from GET /users.
func TestNonAdminCannotListUsers(t *testing.T) {
	userToken := registerAndLogin(t, uniqueEmail("user_nolist"))

	resp := doRequest(t, "GET", "/api/v1/users", nil, userToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

// TestAdminDeleteUser verifies DELETE /users/:id removes the user.
func TestAdminDeleteUser(t *testing.T) {
	_, adminToken := createAdminUser(t, "user_del")

	// Create a user to delete
	victimID, _ := registerGetIDAndLogin(t, uniqueEmail("victim_del"))

	// Delete the user
	resp := doRequest(t, "DELETE", "/api/v1/users/"+victimID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete user, got %d", resp.StatusCode)
	}

	// Verify user is gone — GET should return 404
	resp = doRequest(t, "GET", "/api/v1/users/"+victimID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 on deleted user, got %d", resp.StatusCode)
	}
}

// TestNonAdminCannotDeleteUser verifies regular users cannot delete others.
func TestNonAdminCannotDeleteUser(t *testing.T) {
	_, adminToken := createAdminUser(t, "user_del_nopriv")

	victimID, _ := registerGetIDAndLogin(t, uniqueEmail("victim_nopriv"))
	attackerToken := registerAndLogin(t, uniqueEmail("attacker_del"))
	_ = adminToken

	resp := doRequest(t, "DELETE", "/api/v1/users/"+victimID, nil, attackerToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on non-admin delete, got %d", resp.StatusCode)
	}
}

// TestAdminUpdatePermissions verifies PUT /users/:id/permissions success path.
func TestAdminUpdatePermissions(t *testing.T) {
	_, adminToken := createAdminUser(t, "user_perms")

	targetID, _ := registerGetIDAndLogin(t, uniqueEmail("perms_target"))

	// First get the user's current permissions to find a valid permission_id
	resp := doRequest(t, "GET", "/api/v1/users/"+targetID+"/permissions", nil, adminToken)
	perms := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on get permissions, got %d", resp.StatusCode)
	}
	if len(perms) == 0 {
		t.Skip("no permissions available to update")
	}

	firstPerm := perms[0].(map[string]interface{})
	permID := firstPerm["id"].(string)

	// Update: grant a specific permission
	resp = doRequest(t, "PUT", "/api/v1/users/"+targetID+"/permissions", map[string]interface{}{
		"permissions": []map[string]interface{}{
			{"permission_id": permID, "granted": true},
		},
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update permissions, got %d: %v", resp.StatusCode, data)
	}

	// Verify the permission was granted by re-reading
	resp = doRequest(t, "GET", "/api/v1/users/"+targetID+"/permissions", nil, adminToken)
	updatedPerms := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	found := false
	for _, p := range updatedPerms {
		m := p.(map[string]interface{})
		if m["id"] == permID {
			if granted, ok := m["granted"].(bool); ok && granted {
				found = true
			}
			break
		}
	}
	if !found {
		t.Fatal("expected the permission to be granted after update")
	}
}
