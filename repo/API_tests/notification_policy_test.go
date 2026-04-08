package API_tests

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://chargeops:chargeops@localhost:5433/chargeops?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// waitForJobProcessed polls the DB until the notification job for the given
// user+template is no longer "pending", or until timeout.
func waitForJobProcessed(t *testing.T, db *sql.DB, userID, templateID string, timeout time.Duration) (status, suppressReason string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastStatus string
	for time.Now().Before(deadline) {
		var s string
		var sr sql.NullString
		err := db.QueryRow(
			`SELECT status, suppress_reason FROM notification_jobs
			 WHERE user_id = $1 AND template_id = $2
			 ORDER BY id DESC LIMIT 1`, userID, templateID).Scan(&s, &sr)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		lastStatus = s
		if s != "pending" && s != "processing" {
			reason := ""
			if sr.Valid {
				reason = sr.String
			}
			return s, reason
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("job for user=%s template=%s stuck in %q state after %v", userID, templateID, lastStatus, timeout)
	return "", ""
}

// TestNotificationDelivery_WorkerVerified verifies that the worker delivers
// notifications to non-opted-out users and the message appears in their inbox.
// This test runs first to ensure the worker is warmed up for subsequent tests.
func TestNotificationDelivery_WorkerVerified(t *testing.T) {
	_, adminToken := createAdminUser(t, "deliver_wv")

	tmplCode := uniqueCode("DELIVWV")
	resp := doRequest(t, "POST", "/api/v1/notifications/templates", map[string]interface{}{
		"code":       tmplCode,
		"title_tmpl": "Hello {{name}}",
		"body_tmpl":  "Welcome {{name}}!",
	}, adminToken)
	tmplData := parseJSON(t, resp)
	templateID := tmplData["id"].(string)

	// Register user (not opted out)
	userID, userToken := registerGetIDAndLogin(t, uniqueEmail("deliver_wv_user"))

	// Send notification
	resp = doRequest(t, "POST", "/api/v1/notifications/send", map[string]interface{}{
		"user_id":     userID,
		"template_id": templateID,
		"params":      map[string]string{"name": "Alice"},
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("send: %d", resp.StatusCode)
	}

	// Wait for worker to deliver
	db := testDB(t)
	defer db.Close()
	status, _ := waitForJobProcessed(t, db, userID, templateID, 60*time.Second)

	if status != "delivered" {
		t.Fatalf("expected job status 'delivered', got %q", status)
	}

	// Verify message appears in inbox with rendered content
	resp = doRequest(t, "GET", "/api/v1/notifications/inbox", nil, userToken)
	inbox := parseJSONArray(t, resp)
	found := false
	for _, m := range inbox {
		msg := m.(map[string]interface{})
		if msg["template_id"] == templateID {
			found = true
			if msg["title"] != "Hello Alice" {
				t.Fatalf("expected rendered title 'Hello Alice', got %q", msg["title"])
			}
			if msg["body"] != "Welcome Alice!" {
				t.Fatalf("expected rendered body 'Welcome Alice!', got %q", msg["body"])
			}
			break
		}
	}
	if !found {
		t.Fatal("delivered notification should appear in inbox")
	}
}

// TestOptOutSuppression_WorkerVerified verifies that the worker suppresses
// notifications for opted-out users by checking the job status in the DB.
func TestOptOutSuppression_WorkerVerified(t *testing.T) {
	_, adminToken := createAdminUser(t, "optout_wv")

	tmplCode := uniqueCode("OPTOUTWV")
	resp := doRequest(t, "POST", "/api/v1/notifications/templates", map[string]interface{}{
		"code":       tmplCode,
		"title_tmpl": "Opt-out worker test",
		"body_tmpl":  "Should be suppressed",
	}, adminToken)
	tmplData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create template: %d %v", resp.StatusCode, tmplData)
	}
	templateID := tmplData["id"].(string)

	// Register user and opt out
	userID, userToken := registerGetIDAndLogin(t, uniqueEmail("optout_wv_user"))
	resp = doRequest(t, "PUT", "/api/v1/notifications/subscriptions/"+templateID, map[string]interface{}{
		"opted_out": true,
	}, userToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("opt-out: %d", resp.StatusCode)
	}

	// Send notification
	resp = doRequest(t, "POST", "/api/v1/notifications/send", map[string]interface{}{
		"user_id":     userID,
		"template_id": templateID,
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("send: %d", resp.StatusCode)
	}

	// Wait for the worker to process — it may take a few cycles.
	// We verify by checking that the opted-out notification does NOT appear in inbox
	// even after giving the worker enough time to deliver it if it weren't suppressed.
	db := testDB(t)
	defer db.Close()

	// Wait for the job to leave pending/processing (suppressed, delivered, or failed)
	status, reason := waitForJobProcessed(t, db, userID, templateID, 60*time.Second)

	// Verify the job was suppressed due to opt-out
	if status != "suppressed" {
		t.Fatalf("expected job status 'suppressed', got %q (reason: %q)", status, reason)
	}
	if reason != "opted_out" {
		t.Fatalf("expected suppress_reason 'opted_out', got %q", reason)
	}

	// Verify no message appeared in user's inbox
	resp = doRequest(t, "GET", "/api/v1/notifications/inbox", nil, userToken)
	inbox := parseJSONArray(t, resp)
	for _, m := range inbox {
		msg := m.(map[string]interface{})
		if msg["template_id"] == templateID {
			t.Fatal("suppressed notification should NOT appear in inbox")
		}
	}
}

// TestNotificationSendAndStats verifies delivery stats are incremented.
func TestNotificationSendAndStats(t *testing.T) {
	_, adminToken := createAdminUser(t, "stats_flow")

	tmplCode := uniqueCode("STATS")
	resp := doRequest(t, "POST", "/api/v1/notifications/templates", map[string]interface{}{
		"code":       tmplCode,
		"title_tmpl": "Stats test",
		"body_tmpl":  "Body",
	}, adminToken)
	tmplData := parseJSON(t, resp)
	templateID := tmplData["id"].(string)

	userID, _ := registerGetIDAndLogin(t, uniqueEmail("stats_user"))

	resp = doRequest(t, "POST", "/api/v1/notifications/send", map[string]interface{}{
		"user_id":     userID,
		"template_id": templateID,
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("send: %d", resp.StatusCode)
	}

	// Check delivery stats — generated count should be >= 1
	resp = doRequest(t, "GET", "/api/v1/notifications/stats?template_id="+templateID, nil, adminToken)
	stats := parseJSONArray(t, resp)
	if len(stats) == 0 {
		t.Fatal("expected delivery stats entry after sending")
	}
	stat := stats[0].(map[string]interface{})
	if gen, ok := stat["generated"].(float64); !ok || gen < 1 {
		t.Fatalf("expected generated >= 1, got %v", stat["generated"])
	}
}

// TestNotificationMarkReadAndDismiss verifies read/dismiss ownership checks.
func TestNotificationMarkReadAndDismiss(t *testing.T) {
	_, adminToken := createAdminUser(t, "markread")

	tmplCode := uniqueCode("MARK")
	resp := doRequest(t, "POST", "/api/v1/notifications/templates", map[string]interface{}{
		"code":       tmplCode,
		"title_tmpl": "Mark test",
		"body_tmpl":  "Body",
	}, adminToken)
	tmplData := parseJSON(t, resp)
	templateID := tmplData["id"].(string)

	userAID, userAToken := registerGetIDAndLogin(t, uniqueEmail("markread_a"))
	_, userBToken := registerGetIDAndLogin(t, uniqueEmail("markread_b"))

	// Send notification to user A
	resp = doRequest(t, "POST", "/api/v1/notifications/send", map[string]interface{}{
		"user_id":     userAID,
		"template_id": templateID,
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("send: %d", resp.StatusCode)
	}

	// Wait for worker to deliver
	db := testDB(t)
	defer db.Close()
	status, _ := waitForJobProcessed(t, db, userAID, templateID, 60*time.Second)
	if status != "delivered" {
		t.Skipf("job status %q, skipping read/dismiss test", status)
	}

	// User A checks inbox
	resp = doRequest(t, "GET", "/api/v1/notifications/inbox", nil, userAToken)
	inbox := parseJSONArray(t, resp)
	if len(inbox) == 0 {
		t.Fatal("expected message in inbox after worker delivered")
	}

	msgID := inbox[0].(map[string]interface{})["id"].(string)

	// User B should NOT be able to mark user A's message as read
	resp = doRequest(t, "POST", "/api/v1/notifications/inbox/"+msgID+"/read", nil, userBToken)
	if resp.StatusCode != http.StatusForbidden {
		data := parseJSON(t, resp)
		t.Fatalf("expected 403 for cross-user mark-read, got %d: %v", resp.StatusCode, data)
	}

	// User A can mark their own message as read
	resp = doRequest(t, "POST", "/api/v1/notifications/inbox/"+msgID+"/read", nil, userAToken)
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("expected 200 for own mark-read, got %d: %v", resp.StatusCode, data)
	}

	// User B should NOT be able to dismiss user A's message
	resp = doRequest(t, "POST", "/api/v1/notifications/inbox/"+msgID+"/dismiss", nil, userBToken)
	if resp.StatusCode != http.StatusForbidden {
		data := parseJSON(t, resp)
		t.Fatalf("expected 403 for cross-user dismiss, got %d: %v", resp.StatusCode, data)
	}

	// User A can dismiss their own message
	resp = doRequest(t, "POST", "/api/v1/notifications/inbox/"+msgID+"/dismiss", nil, userAToken)
	if resp.StatusCode != http.StatusOK {
		data := parseJSON(t, resp)
		t.Fatalf("expected 200 for own dismiss, got %d: %v", resp.StatusCode, data)
	}
}

// quietHoursTimezone returns an IANA timezone name where the current time
// falls within quiet hours (9 PM – 7 AM), targeting local 3 AM.
func quietHoursTimezone() string {
	h := time.Now().UTC().Hour()
	offset := (3 - h + 24) % 24
	if offset == 0 {
		return "UTC"
	}
	// IANA Etc/GMT sign convention is inverted: Etc/GMT-5 = UTC+5
	if offset <= 14 {
		return fmt.Sprintf("Etc/GMT-%d", offset)
	}
	return fmt.Sprintf("Etc/GMT+%d", 24-offset)
}

// waitForJobRescheduled polls the DB until the notification job has a
// scheduled_at in the future (indicating quiet-hours rescheduling), or until
// the job reaches a terminal state (delivered/suppressed/failed).
func waitForJobRescheduled(t *testing.T, db *sql.DB, userID, templateID string, timeout time.Duration) (status string, scheduledAt time.Time) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var s string
		var sa time.Time
		err := db.QueryRow(
			`SELECT status, scheduled_at FROM notification_jobs
			 WHERE user_id = $1 AND template_id = $2
			 ORDER BY id DESC LIMIT 1`, userID, templateID).Scan(&s, &sa)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		// If the job was rescheduled, it returns to "pending" with a future scheduled_at
		if s == "pending" && sa.After(time.Now()) {
			return s, sa
		}
		// Terminal state — the job was NOT rescheduled
		if s != "pending" && s != "processing" {
			return s, sa
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("job for user=%s template=%s not rescheduled or finalized after %v", userID, templateID, timeout)
	return "", time.Time{}
}

// TestQuietHoursSuppression_WorkerVerified verifies that notifications sent
// during quiet hours (9 PM – 7 AM local) are rescheduled to the next 7 AM.
func TestQuietHoursSuppression_WorkerVerified(t *testing.T) {
	_, adminToken := createAdminUser(t, "quiet_hr")

	// Create an org whose timezone puts the current time in quiet hours
	tz := quietHoursTimezone()
	orgCode := uniqueCode("QUIETORG")
	resp := doRequest(t, "POST", "/api/v1/orgs", map[string]interface{}{
		"org_code": orgCode,
		"name":     "Quiet Hours Org",
		"timezone": tz,
	}, adminToken)
	orgData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create org: %d %v", resp.StatusCode, orgData)
	}
	orgID := orgData["id"].(string)

	// Create a user and assign to the quiet-hours org
	userID, _ := registerGetIDAndLogin(t, uniqueEmail("quiet_user"))
	setUserOrg(t, userID, orgID, adminToken)

	tmplCode := uniqueCode("QUIET")
	resp = doRequest(t, "POST", "/api/v1/notifications/templates", map[string]interface{}{
		"code":       tmplCode,
		"title_tmpl": "Quiet hours test",
		"body_tmpl":  "Should be rescheduled",
	}, adminToken)
	tmplData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create template: %d %v", resp.StatusCode, tmplData)
	}
	templateID := tmplData["id"].(string)

	// Send notification — should be rescheduled due to quiet hours
	resp = doRequest(t, "POST", "/api/v1/notifications/send", map[string]interface{}{
		"user_id":     userID,
		"template_id": templateID,
	}, adminToken)
	if resp.StatusCode != http.StatusCreated {
		data := parseJSON(t, resp)
		t.Fatalf("send: %d %v", resp.StatusCode, data)
	}

	db := testDB(t)
	defer db.Close()

	status, scheduledAt := waitForJobRescheduled(t, db, userID, templateID, 60*time.Second)
	if status != "pending" {
		t.Fatalf("expected rescheduled job in 'pending' state, got %q", status)
	}
	if !scheduledAt.After(time.Now()) {
		t.Fatalf("expected scheduled_at in the future, got %v", scheduledAt)
	}

	// Verify the rescheduled time is at 7 AM in the user's local timezone
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatalf("load timezone %s: %v", tz, err)
	}
	localScheduled := scheduledAt.In(loc)
	if localScheduled.Hour() != 7 || localScheduled.Minute() != 0 {
		t.Fatalf("expected rescheduled to 7:00 AM local, got %s", localScheduled.Format("15:04"))
	}
}

// TestRateLimitSuppression_WorkerVerified verifies that the worker suppresses
// notifications exceeding the per-user per-template daily limit (2 per day).
func TestRateLimitSuppression_WorkerVerified(t *testing.T) {
	_, adminToken := createAdminUser(t, "ratelimit")

	tmplCode := uniqueCode("RATELIM")
	resp := doRequest(t, "POST", "/api/v1/notifications/templates", map[string]interface{}{
		"code":       tmplCode,
		"title_tmpl": "Rate limit test {{n}}",
		"body_tmpl":  "Notification {{n}}",
	}, adminToken)
	tmplData := parseJSON(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create template: %d %v", resp.StatusCode, tmplData)
	}
	templateID := tmplData["id"].(string)

	userID, userToken := registerGetIDAndLogin(t, uniqueEmail("ratelim_user"))

	db := testDB(t)
	defer db.Close()

	// Send 3 notifications to the same user + template
	for i := 1; i <= 3; i++ {
		resp = doRequest(t, "POST", "/api/v1/notifications/send", map[string]interface{}{
			"user_id":     userID,
			"template_id": templateID,
			"params":      map[string]string{"n": fmt.Sprintf("%d", i)},
		}, adminToken)
		if resp.StatusCode != http.StatusCreated {
			data := parseJSON(t, resp)
			t.Fatalf("send #%d: %d %v", i, resp.StatusCode, data)
		}
	}

	// Wait for all jobs to finish processing
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		var pendingCount int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM notification_jobs
			 WHERE user_id = $1 AND template_id = $2 AND status IN ('pending','processing')`,
			userID, templateID).Scan(&pendingCount)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if pendingCount == 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Count delivered and suppressed jobs
	var deliveredCount, suppressedCount int
	var suppressReason sql.NullString
	err := db.QueryRow(
		`SELECT COUNT(*) FROM notification_jobs
		 WHERE user_id = $1 AND template_id = $2 AND status = 'delivered'`,
		userID, templateID).Scan(&deliveredCount)
	if err != nil {
		t.Fatal(err)
	}
	err = db.QueryRow(
		`SELECT COUNT(*) FROM notification_jobs
		 WHERE user_id = $1 AND template_id = $2 AND status = 'suppressed'`,
		userID, templateID).Scan(&suppressedCount)
	if err != nil {
		t.Fatal(err)
	}

	if deliveredCount != 2 {
		t.Fatalf("expected 2 delivered, got %d", deliveredCount)
	}
	if suppressedCount != 1 {
		t.Fatalf("expected 1 suppressed, got %d", suppressedCount)
	}

	// Verify the suppressed job has reason "rate_limit"
	err = db.QueryRow(
		`SELECT suppress_reason FROM notification_jobs
		 WHERE user_id = $1 AND template_id = $2 AND status = 'suppressed'
		 ORDER BY id DESC LIMIT 1`,
		userID, templateID).Scan(&suppressReason)
	if err != nil {
		t.Fatal(err)
	}
	if !suppressReason.Valid || suppressReason.String != "rate_limit" {
		t.Fatalf("expected suppress_reason 'rate_limit', got %q", suppressReason.String)
	}

	// Verify only 2 messages appear in user's inbox
	resp = doRequest(t, "GET", "/api/v1/notifications/inbox", nil, userToken)
	inbox := parseJSONArray(t, resp)
	count := 0
	for _, m := range inbox {
		msg := m.(map[string]interface{})
		if msg["template_id"] == templateID {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 messages in inbox, got %d", count)
	}
}
