package unit_tests

import (
	"testing"
	"time"

	"github.com/chargeops/api/internal/worker"
)

// These tests validate the production notification worker policy functions
// directly, without requiring database connectivity.

// ---------------------------------------------------------------------------
// ComputeNext7AM — tests the production worker.ComputeNext7AM function
// ---------------------------------------------------------------------------

func TestComputeNext7AM(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	utc := time.UTC

	tests := []struct {
		name    string
		now     time.Time
		loc     *time.Location
		wantDay int
		wantHr  int
	}{
		{
			name:    "at 9 PM same day → next day 7 AM",
			now:     time.Date(2026, 4, 6, 21, 0, 0, 0, utc),
			loc:     utc,
			wantDay: 7,
			wantHr:  7,
		},
		{
			name:    "at 6 AM same day → same day 7 AM",
			now:     time.Date(2026, 4, 6, 6, 0, 0, 0, utc),
			loc:     utc,
			wantDay: 6,
			wantHr:  7,
		},
		{
			name:    "at 7 AM exactly → next day 7 AM",
			now:     time.Date(2026, 4, 6, 7, 0, 0, 0, utc),
			loc:     utc,
			wantDay: 7,
			wantHr:  7,
		},
		{
			name:    "at 11:59 PM → next day 7 AM",
			now:     time.Date(2026, 4, 6, 23, 59, 0, 0, utc),
			loc:     utc,
			wantDay: 7,
			wantHr:  7,
		},
		{
			name:    "at midnight (0 AM) → same day 7 AM",
			now:     time.Date(2026, 4, 6, 0, 0, 0, 0, utc),
			loc:     utc,
			wantDay: 6,
			wantHr:  7,
		},
		{
			name:    "EST timezone at 10 PM → next day 7 AM EST",
			now:     time.Date(2026, 4, 6, 22, 0, 0, 0, est),
			loc:     est,
			wantDay: 7,
			wantHr:  7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := worker.ComputeNext7AM(tc.now, tc.loc)
			if result.Day() != tc.wantDay {
				t.Fatalf("day = %d, want %d", result.Day(), tc.wantDay)
			}
			if result.Hour() != tc.wantHr {
				t.Fatalf("hour = %d, want %d", result.Hour(), tc.wantHr)
			}
			if result.Minute() != 0 || result.Second() != 0 {
				t.Fatalf("expected exactly on the hour, got %v", result)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Quiet hours — tests the production worker.IsQuietHours function
// ---------------------------------------------------------------------------

func TestQuietHoursBoundaries(t *testing.T) {
	tests := []struct {
		hour  int
		quiet bool
	}{
		{0, true},   // midnight
		{1, true},
		{6, true},   // 6 AM
		{7, false},  // 7 AM - quiet hours end
		{8, false},
		{12, false},
		{17, false},
		{20, false}, // 8 PM
		{21, true},  // 9 PM - quiet hours start
		{22, true},
		{23, true},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			got := worker.IsQuietHours(tc.hour)
			if got != tc.quiet {
				t.Fatalf("hour %d: quiet=%v, want %v", tc.hour, got, tc.quiet)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RenderTemplate — tests the production worker.RenderTemplate function
// ---------------------------------------------------------------------------

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name   string
		tmpl   string
		params string
		want   string
	}{
		{
			name:   "single substitution",
			tmpl:   "Hello {{name}}!",
			params: `{"name":"Alice"}`,
			want:   "Hello Alice!",
		},
		{
			name:   "multiple substitutions",
			tmpl:   "{{greeting}} {{name}}, your order {{order_id}} is ready",
			params: `{"greeting":"Hi","name":"Bob","order_id":"ORD-42"}`,
			want:   "Hi Bob, your order ORD-42 is ready",
		},
		{
			name:   "no params returns template unchanged",
			tmpl:   "Hello {{name}}!",
			params: "",
			want:   "Hello {{name}}!",
		},
		{
			name:   "invalid JSON returns template unchanged",
			tmpl:   "Hello {{name}}!",
			params: "not json",
			want:   "Hello {{name}}!",
		},
		{
			name:   "no placeholders in template",
			tmpl:   "Hello world!",
			params: `{"name":"Alice"}`,
			want:   "Hello world!",
		},
		{
			name:   "unmatched placeholder preserved",
			tmpl:   "Hello {{name}}, your code is {{code}}",
			params: `{"name":"Alice"}`,
			want:   "Hello Alice, your code is {{code}}",
		},
		{
			name:   "empty params object",
			tmpl:   "Hello {{name}}!",
			params: `{}`,
			want:   "Hello {{name}}!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := worker.RenderTemplate(tc.tmpl, []byte(tc.params))
			if got != tc.want {
				t.Fatalf("RenderTemplate(%q, %q) = %q, want %q", tc.tmpl, tc.params, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Rate limit window computation
// ---------------------------------------------------------------------------

func TestRateLimitWindowComputation(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	jst, _ := time.LoadLocation("Asia/Tokyo")

	tests := []struct {
		name      string
		now       time.Time
		loc       *time.Location
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "EST midday → midnight-to-midnight EST",
			now:       time.Date(2026, 4, 6, 14, 0, 0, 0, est),
			loc:       est,
			wantStart: time.Date(2026, 4, 6, 0, 0, 0, 0, est),
			wantEnd:   time.Date(2026, 4, 7, 0, 0, 0, 0, est),
		},
		{
			name:      "JST evening → same local day",
			now:       time.Date(2026, 4, 6, 22, 0, 0, 0, jst),
			loc:       jst,
			wantStart: time.Date(2026, 4, 6, 0, 0, 0, 0, jst),
			wantEnd:   time.Date(2026, 4, 7, 0, 0, 0, 0, jst),
		},
		{
			name:      "UTC at midnight → covers full UTC day",
			now:       time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
			loc:       time.UTC,
			wantStart: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			localNow := tc.now.In(tc.loc)
			y, m, d := localNow.Date()
			localDay := time.Date(y, m, d, 0, 0, 0, 0, tc.loc)
			nextDay := localDay.AddDate(0, 0, 1)

			if !localDay.Equal(tc.wantStart) {
				t.Fatalf("window start = %v, want %v", localDay, tc.wantStart)
			}
			if !nextDay.Equal(tc.wantEnd) {
				t.Fatalf("window end = %v, want %v", nextDay, tc.wantEnd)
			}

			startUTC := localDay.UTC()
			endUTC := nextDay.UTC()
			if !startUTC.Before(endUTC) {
				t.Fatalf("UTC start (%v) should be before UTC end (%v)", startUTC, endUTC)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST transition: rate limit window must still cover exactly one local day
// ---------------------------------------------------------------------------

func TestRateLimitWindowDSTTransition(t *testing.T) {
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York timezone not available")
	}

	// March 8, 2026: US DST spring forward (2 AM → 3 AM)
	localNow := time.Date(2026, 3, 8, 14, 0, 0, 0, est)
	y, m, d := localNow.Date()
	localDay := time.Date(y, m, d, 0, 0, 0, 0, est)
	nextDay := localDay.AddDate(0, 0, 1)

	windowDuration := nextDay.Sub(localDay)
	if windowDuration != 23*time.Hour {
		t.Fatalf("DST spring-forward day should be 23h, got %v", windowDuration)
	}

	// November 1, 2026: US DST fall back (2 AM → 1 AM)
	localNow = time.Date(2026, 11, 1, 14, 0, 0, 0, est)
	y, m, d = localNow.Date()
	localDay = time.Date(y, m, d, 0, 0, 0, 0, est)
	nextDay = localDay.AddDate(0, 0, 1)

	windowDuration = nextDay.Sub(localDay)
	if windowDuration != 25*time.Hour {
		t.Fatalf("DST fall-back day should be 25h, got %v", windowDuration)
	}
}
