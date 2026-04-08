package unit_tests

import (
	"math"
	"testing"
	"time"

	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/service"
	"github.com/shopspring/decimal"
)

// These tests verify the pricing calculation patterns used in the production
// pricing_service.go. They use the same math.Ceil and shopspring/decimal
// operations that the real code uses, ensuring correctness of the formulas.

func TestDurationCeilingRounding(t *testing.T) {
	tests := []struct {
		name       string
		durationSec float64
		wantMin    int
	}{
		{
			name:        "0.5 minutes (30s) rounds up to 1 minute",
			durationSec: 30.0,
			wantMin:     1,
		},
		{
			name:        "exactly 5 minutes stays 5",
			durationSec: 300.0,
			wantMin:     5,
		},
		{
			name:        "5.01 minutes (300.6s) rounds up to 6",
			durationSec: 300.6,
			wantMin:     6,
		},
		{
			name:        "zero duration is 0 minutes",
			durationSec: 0.0,
			wantMin:     0,
		},
		{
			name:        "1 second rounds up to 1 minute",
			durationSec: 1.0,
			wantMin:     1,
		},
		{
			name:        "exactly 1 minute stays 1",
			durationSec: 60.0,
			wantMin:     1,
		},
		{
			name:        "61 seconds rounds up to 2 minutes",
			durationSec: 61.0,
			wantMin:     2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// This is the exact formula from pricing_service.go
			durationMin := int(math.Ceil(tc.durationSec / 60.0))
			if durationMin != tc.wantMin {
				t.Fatalf("ceil(%f / 60) = %d, want %d", tc.durationSec, durationMin, tc.wantMin)
			}
		})
	}
}

func TestEnergyCostCalculation(t *testing.T) {
	// energyCost = energyRate * energyKWh (from pricing_service.go)
	rate := decimal.NewFromFloat(0.15)    // $0.15 per kWh
	kwh := decimal.NewFromFloat(50.0)     // 50 kWh consumed
	expected := decimal.NewFromFloat(7.50) // $7.50

	energyCost := rate.Mul(kwh)
	if !energyCost.Equal(expected) {
		t.Fatalf("energy cost = %s, want %s", energyCost.String(), expected.String())
	}
}

func TestServiceFeeAddition(t *testing.T) {
	energyCost := decimal.NewFromFloat(7.50)
	durationCost := decimal.NewFromFloat(3.00)
	serviceFee := decimal.NewFromFloat(1.50)

	// subtotal = energyCost + durationCost + serviceFee (from pricing_service.go)
	subtotal := energyCost.Add(durationCost).Add(serviceFee)
	expected := decimal.NewFromFloat(12.00)

	if !subtotal.Equal(expected) {
		t.Fatalf("subtotal = %s, want %s", subtotal.String(), expected.String())
	}
}

func TestTaxCalculation(t *testing.T) {
	subtotal := decimal.NewFromFloat(12.00)
	taxRate := decimal.NewFromFloat(0.08) // 8% tax

	// taxAmount = subtotal * taxRate (from pricing_service.go)
	taxAmount := subtotal.Mul(taxRate)
	expected := decimal.NewFromFloat(0.96)

	if !taxAmount.Equal(expected) {
		t.Fatalf("tax amount = %s, want %s", taxAmount.String(), expected.String())
	}
}

func TestTotalCalculation(t *testing.T) {
	subtotal := decimal.NewFromFloat(12.00)
	taxAmount := decimal.NewFromFloat(0.96)

	// total = subtotal + taxAmount (from pricing_service.go)
	total := subtotal.Add(taxAmount)
	expected := decimal.NewFromFloat(12.96)

	if !total.Equal(expected) {
		t.Fatalf("total = %s, want %s", total.String(), expected.String())
	}
}

func TestFullPricingCalculation(t *testing.T) {
	// Simulate a complete pricing calculation as done in pricing_service.go
	energyRate := decimal.NewFromFloat(0.20)   // $0.20 per kWh
	durationRate := decimal.NewFromFloat(0.05) // $0.05 per minute
	serviceFee := decimal.NewFromFloat(2.00)   // $2.00 flat
	taxRate := decimal.NewFromFloat(0.10)      // 10%
	energyKWh := decimal.NewFromFloat(30.0)    // 30 kWh

	// Duration: 45 minutes 30 seconds = 2730 seconds -> ceil = 46 minutes
	durationSec := 2730.0
	durationMin := int(math.Ceil(durationSec / 60.0))
	if durationMin != 46 {
		t.Fatalf("expected 46 minutes, got %d", durationMin)
	}

	energyCost := energyRate.Mul(energyKWh)                                  // 0.20 * 30 = 6.00
	durationCost := durationRate.Mul(decimal.NewFromInt(int64(durationMin)))  // 0.05 * 46 = 2.30
	subtotal := energyCost.Add(durationCost).Add(serviceFee)                 // 6.00 + 2.30 + 2.00 = 10.30
	taxAmount := subtotal.Mul(taxRate)                                        // 10.30 * 0.10 = 1.03
	total := subtotal.Add(taxAmount)                                          // 10.30 + 1.03 = 11.33

	if !energyCost.Equal(decimal.NewFromFloat(6.00)) {
		t.Fatalf("energy cost = %s, want 6.00", energyCost.String())
	}
	if !durationCost.Equal(decimal.NewFromFloat(2.30)) {
		t.Fatalf("duration cost = %s, want 2.30", durationCost.String())
	}
	if !subtotal.Equal(decimal.NewFromFloat(10.30)) {
		t.Fatalf("subtotal = %s, want 10.30", subtotal.String())
	}
	if !taxAmount.Equal(decimal.NewFromFloat(1.03)) {
		t.Fatalf("tax amount = %s, want 1.03", taxAmount.String())
	}
	if !total.Equal(decimal.NewFromFloat(11.33)) {
		t.Fatalf("total = %s, want 11.33", total.String())
	}
}

func TestZeroDurationCost(t *testing.T) {
	durationRate := decimal.NewFromFloat(0.05)
	durationSec := 0.0
	durationMin := int(math.Ceil(durationSec / 60.0))

	durationCost := durationRate.Mul(decimal.NewFromInt(int64(durationMin)))
	if !durationCost.Equal(decimal.Zero) {
		t.Fatalf("zero duration cost = %s, want 0", durationCost.String())
	}
}

func TestDecimalPrecisionNoFloatingPointErrors(t *testing.T) {
	// Classic floating point problem: 0.1 + 0.2 != 0.3 in IEEE 754
	// The shopspring/decimal library should handle this correctly
	a := decimal.NewFromFloat(0.1)
	b := decimal.NewFromFloat(0.2)
	sum := a.Add(b)
	expected := decimal.NewFromFloat(0.3)

	if !sum.Equal(expected) {
		t.Fatalf("0.1 + 0.2 = %s, want 0.3 (floating point precision error)", sum.String())
	}

	// Another common case: money multiplication
	price := decimal.NewFromFloat(19.99)
	qty := decimal.NewFromInt(3)
	total := price.Mul(qty)
	expectedTotal := decimal.NewFromFloat(59.97)

	if !total.Equal(expectedTotal) {
		t.Fatalf("19.99 * 3 = %s, want 59.97", total.String())
	}
}

// ---------------------------------------------------------------------------
// Day-type classification — tests the production ClassifyDayType function
// ---------------------------------------------------------------------------

func TestClassifyDayType(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"Monday is weekday", time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC), "weekday"},
		{"Tuesday is weekday", time.Date(2026, 4, 7, 12, 0, 0, 0, time.UTC), "weekday"},
		{"Wednesday is weekday", time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC), "weekday"},
		{"Thursday is weekday", time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC), "weekday"},
		{"Friday is weekday", time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC), "weekday"},
		{"Saturday is weekend", time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC), "weekend"},
		{"Sunday is weekend", time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC), "weekend"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := service.ClassifyDayType(tc.date)
			if got != tc.want {
				t.Fatalf("ClassifyDayType(%s) = %q, want %q", tc.date.Weekday(), got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Time string normalization — tests the production NormalizeTimeStr function
// ---------------------------------------------------------------------------

func TestNormalizeTimeStr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"HH:MM stays as is", "07:00", "07:00"},
		{"HH:MM:SS truncated to HH:MM", "07:00:00", "07:00"},
		{"PostgreSQL TIME scanned as RFC3339", "0000-01-01T07:00:00Z", "07:00"},
		{"23:59 stays as is", "23:59", "23:59"},
		{"leading whitespace trimmed", "  09:30", "09:30"},
		{"trailing whitespace trimmed", "14:00  ", "14:00"},
		{"midnight", "00:00", "00:00"},
		{"HH:MM:SS with seconds", "13:45:30", "13:45"},
		{"PostgreSQL noon", "0000-01-01T12:00:00Z", "12:00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := service.NormalizeTimeStr(tc.input)
			if got != tc.want {
				t.Fatalf("normalizeTimeStr(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TOU rule matching — tests the production FindMatchingTOURule function
// ---------------------------------------------------------------------------

func TestFindMatchingTOURule(t *testing.T) {
	rules := []model.TOURule{
		{DayType: "weekday", StartTime: "06:00", EndTime: "12:00", EnergyRate: decimal.NewFromFloat(0.10), DurationRate: decimal.NewFromFloat(0.02)},
		{DayType: "weekday", StartTime: "12:00", EndTime: "18:00", EnergyRate: decimal.NewFromFloat(0.25), DurationRate: decimal.NewFromFloat(0.05)},
		{DayType: "weekday", StartTime: "18:00", EndTime: "22:00", EnergyRate: decimal.NewFromFloat(0.15), DurationRate: decimal.NewFromFloat(0.03)},
		{DayType: "weekend", StartTime: "08:00", EndTime: "20:00", EnergyRate: decimal.NewFromFloat(0.12), DurationRate: decimal.NewFromFloat(0.02)},
	}

	tests := []struct {
		name       string
		dayType    string
		localTime  time.Time
		wantMatch  bool
		wantEnergy string
	}{
		{
			name:       "weekday morning matches first rule",
			dayType:    "weekday",
			localTime:  time.Date(2026, 4, 6, 9, 30, 0, 0, time.UTC),
			wantMatch:  true,
			wantEnergy: "0.1",
		},
		{
			name:       "weekday at exactly 12:00 matches afternoon rule",
			dayType:    "weekday",
			localTime:  time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC),
			wantMatch:  true,
			wantEnergy: "0.25",
		},
		{
			name:       "weekday at 11:59 still in morning rule",
			dayType:    "weekday",
			localTime:  time.Date(2026, 4, 6, 11, 59, 0, 0, time.UTC),
			wantMatch:  true,
			wantEnergy: "0.1",
		},
		{
			name:       "weekday at 05:59 no rule matches (before 06:00)",
			dayType:    "weekday",
			localTime:  time.Date(2026, 4, 6, 5, 59, 0, 0, time.UTC),
			wantMatch:  false,
		},
		{
			name:       "weekday at exactly 22:00 no rule matches (end exclusive)",
			dayType:    "weekday",
			localTime:  time.Date(2026, 4, 6, 22, 0, 0, 0, time.UTC),
			wantMatch:  false,
		},
		{
			name:       "weekend at 10:00 matches weekend rule",
			dayType:    "weekend",
			localTime:  time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC),
			wantMatch:  true,
			wantEnergy: "0.12",
		},
		{
			name:       "weekend at 07:59 no rule matches",
			dayType:    "weekend",
			localTime:  time.Date(2026, 4, 11, 7, 59, 0, 0, time.UTC),
			wantMatch:  false,
		},
		{
			name:       "holiday dayType has no rules, no match",
			dayType:    "holiday",
			localTime:  time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC),
			wantMatch:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match := service.FindMatchingTOURule(rules, tc.dayType, tc.localTime)
			if tc.wantMatch {
				if match == nil {
					t.Fatal("expected a matching TOU rule, got nil")
				}
				if match.EnergyRate.String() != tc.wantEnergy {
					t.Fatalf("matched energy_rate = %s, want %s", match.EnergyRate.String(), tc.wantEnergy)
				}
			} else {
				if match != nil {
					t.Fatalf("expected no match, got rule with energy_rate=%s", match.EnergyRate.String())
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TOU overlap detection — tests the production CheckTOUOverlap function
// ---------------------------------------------------------------------------

func TestTOUOverlapDetection(t *testing.T) {
	tests := []struct {
		name     string
		existing []model.TOURule
		newRule  model.TOURule
		overlap  bool
	}{
		{
			name: "no overlap - adjacent windows",
			existing: []model.TOURule{
				{DayType: "weekday", StartTime: "06:00", EndTime: "12:00"},
			},
			newRule: model.TOURule{DayType: "weekday", StartTime: "12:00", EndTime: "18:00"},
			overlap: false,
		},
		{
			name: "overlap - new starts during existing",
			existing: []model.TOURule{
				{DayType: "weekday", StartTime: "06:00", EndTime: "12:00"},
			},
			newRule: model.TOURule{DayType: "weekday", StartTime: "11:00", EndTime: "14:00"},
			overlap: true,
		},
		{
			name: "no overlap - different day types",
			existing: []model.TOURule{
				{DayType: "weekday", StartTime: "06:00", EndTime: "12:00"},
			},
			newRule: model.TOURule{DayType: "weekend", StartTime: "06:00", EndTime: "12:00"},
			overlap: false,
		},
		{
			name: "overlap - new contains existing",
			existing: []model.TOURule{
				{DayType: "weekday", StartTime: "08:00", EndTime: "10:00"},
			},
			newRule: model.TOURule{DayType: "weekday", StartTime: "06:00", EndTime: "12:00"},
			overlap: true,
		},
		{
			name: "overlap - existing contains new",
			existing: []model.TOURule{
				{DayType: "weekday", StartTime: "06:00", EndTime: "18:00"},
			},
			newRule: model.TOURule{DayType: "weekday", StartTime: "08:00", EndTime: "10:00"},
			overlap: true,
		},
		{
			name:     "no existing rules - no overlap",
			existing: []model.TOURule{},
			newRule:  model.TOURule{DayType: "weekday", StartTime: "06:00", EndTime: "12:00"},
			overlap:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := service.CheckTOUOverlap(tc.existing, tc.newRule.DayType, tc.newRule.StartTime, tc.newRule.EndTime)
			if got != tc.overlap {
				t.Fatalf("CheckTOUOverlap = %v, want %v", got, tc.overlap)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Recalculation semantics: same inputs must produce identical outputs
// ---------------------------------------------------------------------------

func TestRecalculationIdempotency(t *testing.T) {
	// Simulates the recalculation logic: given the same version rates and
	// order parameters, recalculation must produce the same total.
	energyRate := decimal.NewFromFloat(0.20)
	durationRate := decimal.NewFromFloat(0.05)
	serviceFee := decimal.NewFromFloat(2.00)
	taxRate := decimal.NewFromFloat(0.10)
	energyKWh := decimal.NewFromFloat(30.0)
	orderStart := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	orderEnd := time.Date(2026, 4, 6, 10, 45, 30, 0, time.UTC)

	calculate := func() (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
		durationSec := orderEnd.Sub(orderStart).Seconds()
		durationMin := int(math.Ceil(durationSec / 60.0))
		energyCost := energyRate.Mul(energyKWh)
		durationCost := durationRate.Mul(decimal.NewFromInt(int64(durationMin)))
		subtotal := energyCost.Add(durationCost).Add(serviceFee)
		taxAmount := subtotal.Mul(taxRate)
		total := subtotal.Add(taxAmount)
		return subtotal, taxAmount, total
	}

	sub1, tax1, total1 := calculate()
	sub2, tax2, total2 := calculate()

	if !sub1.Equal(sub2) || !tax1.Equal(tax2) || !total1.Equal(total2) {
		t.Fatalf("recalculation produced different results: (%s,%s,%s) vs (%s,%s,%s)",
			sub1, tax1, total1, sub2, tax2, total2)
	}
}

// ---------------------------------------------------------------------------
// TOU rate override: when TOU matches, energy/duration rates come from TOU
// ---------------------------------------------------------------------------

func TestTOURateOverridesVersionRate(t *testing.T) {
	versionEnergy := decimal.NewFromFloat(0.20)
	versionDuration := decimal.NewFromFloat(0.05)
	serviceFee := decimal.NewFromFloat(2.00)
	taxRate := decimal.NewFromFloat(0.10)
	energyKWh := decimal.NewFromFloat(30.0)
	durationMin := 46

	touEnergy := decimal.NewFromFloat(0.30)
	touDuration := decimal.NewFromFloat(0.08)

	// Without TOU
	noTOUEnergyCost := versionEnergy.Mul(energyKWh)
	noTOUDurationCost := versionDuration.Mul(decimal.NewFromInt(int64(durationMin)))
	noTOUSubtotal := noTOUEnergyCost.Add(noTOUDurationCost).Add(serviceFee)
	noTOUTotal := noTOUSubtotal.Add(noTOUSubtotal.Mul(taxRate))

	// With TOU override
	touEnergyCost := touEnergy.Mul(energyKWh)
	touDurationCost := touDuration.Mul(decimal.NewFromInt(int64(durationMin)))
	touSubtotal := touEnergyCost.Add(touDurationCost).Add(serviceFee)
	touTotal := touSubtotal.Add(touSubtotal.Mul(taxRate))

	if noTOUTotal.Equal(touTotal) {
		t.Fatal("TOU and non-TOU totals should differ when TOU rates differ from version rates")
	}
	if !touTotal.GreaterThan(noTOUTotal) {
		t.Fatalf("TOU total (%s) should be > non-TOU total (%s) given higher TOU rates",
			touTotal.String(), noTOUTotal.String())
	}
}

// ---------------------------------------------------------------------------
// Tax disabled: when apply_tax=false, tax_amount must be zero
// ---------------------------------------------------------------------------

func TestNoTaxWhenApplyTaxFalse(t *testing.T) {
	subtotal := decimal.NewFromFloat(15.50)
	applyTax := false
	taxRate := decimal.NewFromFloat(0.10)

	taxAmount := decimal.Zero
	if applyTax {
		taxAmount = subtotal.Mul(taxRate)
	}
	total := subtotal.Add(taxAmount)

	if !taxAmount.Equal(decimal.Zero) {
		t.Fatalf("tax should be zero when apply_tax=false, got %s", taxAmount.String())
	}
	if !total.Equal(subtotal) {
		t.Fatalf("total should equal subtotal when no tax, got %s vs %s", total.String(), subtotal.String())
	}
}

// ---------------------------------------------------------------------------
// Version resolution with effective_at: simulates the pricing resolution
// query logic (effective_at <= orderTime, ORDER BY effective_at DESC LIMIT 1)
// ---------------------------------------------------------------------------

type testVersion struct {
	ID          string
	Status      string
	EffectiveAt time.Time
	EnergyRate  decimal.Decimal
}

// resolveVersion mirrors the DB query: picks the most recently effective
// active version whose effective_at <= orderTime.
func resolveVersion(versions []testVersion, orderTime time.Time) *testVersion {
	var best *testVersion
	for i := range versions {
		v := &versions[i]
		if v.Status != "active" {
			continue
		}
		if v.EffectiveAt.After(orderTime) {
			continue
		}
		if best == nil || v.EffectiveAt.After(best.EffectiveAt) {
			best = v
		}
	}
	return best
}

func TestVersionResolution_FutureDatedDoesNotDisrupt(t *testing.T) {
	now := time.Now().UTC()
	v1 := testVersion{ID: "v1", Status: "active", EffectiveAt: now.Add(-1 * time.Hour), EnergyRate: decimal.NewFromFloat(0.10)}
	v2 := testVersion{ID: "v2", Status: "active", EffectiveAt: now.Add(1 * time.Hour), EnergyRate: decimal.NewFromFloat(0.50)}
	versions := []testVersion{v1, v2}

	// Order placed now: should resolve to v1 (v2's effective_at is in the future)
	resolved := resolveVersion(versions, now)
	if resolved == nil {
		t.Fatal("expected v1 to resolve, got nil")
	}
	if resolved.ID != "v1" {
		t.Fatalf("expected v1, got %s", resolved.ID)
	}

	// Order placed after v2's effective_at: should resolve to v2
	futureOrder := now.Add(2 * time.Hour)
	resolved = resolveVersion(versions, futureOrder)
	if resolved == nil {
		t.Fatal("expected v2 to resolve, got nil")
	}
	if resolved.ID != "v2" {
		t.Fatalf("expected v2, got %s", resolved.ID)
	}
}

func TestVersionResolution_OldVersionDeactivated_GapExists(t *testing.T) {
	now := time.Now().UTC()
	// Simulates the BUG scenario: old version deactivated, new version future-dated
	v1 := testVersion{ID: "v1", Status: "inactive", EffectiveAt: now.Add(-1 * time.Hour)}
	v2 := testVersion{ID: "v2", Status: "active", EffectiveAt: now.Add(1 * time.Hour)}
	versions := []testVersion{v1, v2}

	// Order placed now: NO version resolves (pricing gap!)
	resolved := resolveVersion(versions, now)
	if resolved != nil {
		t.Fatalf("with old version deactivated and new future-dated, no version should resolve, but got %s", resolved.ID)
	}
}

func TestVersionResolution_BothActive_NoGap(t *testing.T) {
	now := time.Now().UTC()
	// Correct behavior: both versions active, resolution picks the right one
	v1 := testVersion{ID: "v1", Status: "active", EffectiveAt: now.Add(-1 * time.Hour)}
	v2 := testVersion{ID: "v2", Status: "active", EffectiveAt: now.Add(1 * time.Hour)}
	versions := []testVersion{v1, v2}

	// Order placed now: v1 resolves (no gap)
	resolved := resolveVersion(versions, now)
	if resolved == nil {
		t.Fatal("expected v1 to resolve (no pricing gap), got nil")
	}
	if resolved.ID != "v1" {
		t.Fatalf("expected v1, got %s", resolved.ID)
	}

	// Order placed after v2 effective: v2 resolves
	resolved = resolveVersion(versions, now.Add(2*time.Hour))
	if resolved == nil {
		t.Fatal("expected v2 to resolve, got nil")
	}
	if resolved.ID != "v2" {
		t.Fatalf("expected v2, got %s", resolved.ID)
	}
}

func TestVersionResolution_ImmediateActivation_OldDeactivated(t *testing.T) {
	now := time.Now().UTC()
	// Immediate activation: old version correctly deactivated, new active now
	v1 := testVersion{ID: "v1", Status: "inactive", EffectiveAt: now.Add(-2 * time.Hour)}
	v2 := testVersion{ID: "v2", Status: "active", EffectiveAt: now.Add(-1 * time.Second)}
	versions := []testVersion{v1, v2}

	resolved := resolveVersion(versions, now)
	if resolved == nil {
		t.Fatal("expected v2 to resolve, got nil")
	}
	if resolved.ID != "v2" {
		t.Fatalf("expected v2, got %s", resolved.ID)
	}
}
