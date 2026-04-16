package API_tests

import (
	"net/http"
	"testing"
)

func TestAdminStationDeviceCRUD(t *testing.T) {
	_, adminToken := createAdminUser(t, "st_crud")

	// 1. Create org
	orgID := adminCreateOrg(t, adminToken, "storg")

	// 2. Create station
	resp := doRequest(t, "POST", "/api/v1/stations", map[string]interface{}{
		"org_id":   orgID,
		"name":     "Station 1",
		"timezone": "UTC",
	}, adminToken)
	data := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create station, got %d: %v", resp.StatusCode, data)
	}
	stationID := data["id"].(string)

	// 3. GET station
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on GET station, got %d: %v", resp.StatusCode, data)
	}
	if data["name"] != "Station 1" {
		t.Fatalf("expected name 'Station 1', got %v", data["name"])
	}

	// 4. PUT station (update name)
	resp = doRequest(t, "PUT", "/api/v1/stations/"+stationID, map[string]interface{}{
		"name": "Station 1 Updated",
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on PUT station, got %d: %v", resp.StatusCode, data)
	}
	if data["name"] != "Station 1 Updated" {
		t.Fatalf("expected updated name, got %v", data["name"])
	}

	// 5. Create device
	deviceCode := uniqueCode("DEV")
	resp = doRequest(t, "POST", "/api/v1/stations/"+stationID+"/devices", map[string]interface{}{
		"device_code": deviceCode,
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create device, got %d: %v", resp.StatusCode, data)
	}
	deviceID := data["id"].(string)

	// 6. List devices
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID+"/devices", nil, adminToken)
	devices := parseJSONArray(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list devices, got %d", resp.StatusCode)
	}
	if len(devices) < 1 {
		t.Fatal("expected at least 1 device")
	}

	// 7. Update device
	resp = doRequest(t, "PUT", "/api/v1/devices/"+deviceID, map[string]interface{}{
		"status": "inactive",
	}, adminToken)
	data = parseJSON(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update device, got %d: %v", resp.StatusCode, data)
	}

	// 8. Delete device
	resp = doRequest(t, "DELETE", "/api/v1/devices/"+deviceID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete device, got %d", resp.StatusCode)
	}

	// 9. Delete station (DELETE /api/v1/stations/:id)
	resp = doRequest(t, "DELETE", "/api/v1/stations/"+stationID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete station, got %d", resp.StatusCode)
	}

	// Confirm station is gone
	resp = doRequest(t, "GET", "/api/v1/stations/"+stationID, nil, adminToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 on deleted station, got %d", resp.StatusCode)
	}
}
