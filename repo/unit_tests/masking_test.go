package unit_tests

import (
	"strings"
	"testing"

	"github.com/chargeops/api/internal/masking"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/service"
	"github.com/google/uuid"
)

func TestMaskTaxID(t *testing.T) {
	tests := []struct {
		name  string
		taxID string
		role  string
		want  string
	}{
		{
			name:  "admin role returns full value",
			taxID: "123-45-6789",
			role:  "administrator",
			want:  "123-45-6789",
		},
		{
			name:  "user role shows last 4 chars only",
			taxID: "123-45-6789",
			role:  "user",
			want:  "*******6789",
		},
		{
			name:  "merchant role shows last 4 chars only",
			taxID: "123-45-6789",
			role:  "merchant",
			want:  "*******6789",
		},
		{
			name:  "short value (<=4) returns ****",
			taxID: "1234",
			role:  "user",
			want:  "****",
		},
		{
			name:  "very short value (3 chars) returns ****",
			taxID: "abc",
			role:  "user",
			want:  "****",
		},
		{
			name:  "empty string returns empty string",
			taxID: "",
			role:  "user",
			want:  "",
		},
		{
			name:  "empty string with admin returns empty string",
			taxID: "",
			role:  "administrator",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := masking.MaskTaxID(tc.taxID, tc.role)
			if got != tc.want {
				t.Fatalf("MaskTaxID(%q, %q) = %q, want %q", tc.taxID, tc.role, got, tc.want)
			}
		})
	}
}

func TestMaskAddress(t *testing.T) {
	tests := []struct {
		name string
		addr string
		role string
		want string
	}{
		{
			name: "admin role returns full value",
			addr: "123 Main Street, Apt 4",
			role: "administrator",
			want: "123 Main Street, Apt 4",
		},
		{
			name: "user role returns ****",
			addr: "123 Main Street, Apt 4",
			role: "user",
			want: "****",
		},
		{
			name: "merchant role returns ****",
			addr: "456 Oak Ave",
			role: "merchant",
			want: "****",
		},
		{
			name: "empty string returns empty string",
			addr: "",
			role: "user",
			want: "",
		},
		{
			name: "empty string with admin returns empty string",
			addr: "",
			role: "administrator",
			want: "",
		},
		{
			name: "multi-line address masked for user",
			addr: "123 Main Street\nApt 4B\nNew York, NY 10001",
			role: "user",
			want: "****",
		},
		{
			name: "multi-line address visible for admin",
			addr: "123 Main Street\nApt 4B\nNew York, NY 10001",
			role: "administrator",
			want: "123 Main Street\nApt 4B\nNew York, NY 10001",
		},
		{
			name: "guest role returns ****",
			addr: "456 Oak Ave",
			role: "guest",
			want: "****",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := masking.MaskAddress(tc.addr, tc.role)
			if got != tc.want {
				t.Fatalf("MaskAddress(%q, %q) = %q, want %q", tc.addr, tc.role, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Handler-level masking: MaskOrg produces correctly masked OrgResponse
// ---------------------------------------------------------------------------

func TestMaskOrgResponse(t *testing.T) {
	taxID := "123-45-6789"
	address := "742 Evergreen Terrace, Springfield"
	org := &model.Organization{
		ID:      uuid.New(),
		OrgCode: "TEST-ORG",
		Name:    "Test Organization",
		TaxID:   &taxID,
		Address: &address,
	}

	tests := []struct {
		name        string
		role        string
		wantTaxID   string
		wantAddress string
	}{
		{
			name:        "admin sees full tax_id and address",
			role:        "administrator",
			wantTaxID:   "123-45-6789",
			wantAddress: "742 Evergreen Terrace, Springfield",
		},
		{
			name:        "user sees masked tax_id and address",
			role:        "user",
			wantTaxID:   "*******6789",
			wantAddress: "****",
		},
		{
			name:        "merchant sees masked tax_id and address",
			role:        "merchant",
			wantTaxID:   "*******6789",
			wantAddress: "****",
		},
		{
			name:        "guest sees masked tax_id and address",
			role:        "guest",
			wantTaxID:   "*******6789",
			wantAddress: "****",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := service.MaskOrg(org, tc.role)
			if resp.TaxID != tc.wantTaxID {
				t.Fatalf("MaskOrg tax_id = %q, want %q", resp.TaxID, tc.wantTaxID)
			}
			if resp.Address != tc.wantAddress {
				t.Fatalf("MaskOrg address = %q, want %q", resp.Address, tc.wantAddress)
			}
			// Verify non-sensitive fields are unmasked
			if resp.OrgCode != org.OrgCode {
				t.Fatalf("OrgCode should not be masked: got %q, want %q", resp.OrgCode, org.OrgCode)
			}
			if resp.Name != org.Name {
				t.Fatalf("Name should not be masked: got %q, want %q", resp.Name, org.Name)
			}
		})
	}
}

func TestMaskOrgNilFields(t *testing.T) {
	org := &model.Organization{
		ID:      uuid.New(),
		OrgCode: "NO-PII",
		Name:    "No PII Org",
		TaxID:   nil,
		Address: nil,
	}
	resp := service.MaskOrg(org, "user")
	if resp.TaxID != "" {
		t.Fatalf("nil tax_id should produce empty string, got %q", resp.TaxID)
	}
	if resp.Address != "" {
		t.Fatalf("nil address should produce empty string, got %q", resp.Address)
	}
}

// ---------------------------------------------------------------------------
// Handler-level masking: MaskSupplier mutates supplier in place
// ---------------------------------------------------------------------------

func TestMaskSupplierResponse(t *testing.T) {
	tests := []struct {
		name        string
		role        string
		wantTaxID   string
		wantAddress string
	}{
		{
			name:        "admin sees full values",
			role:        "administrator",
			wantTaxID:   "98-7654321",
			wantAddress: "100 Supplier Ave",
		},
		{
			name:        "user sees masked values",
			role:        "user",
			wantTaxID:   "******4321",
			wantAddress: "****",
		},
		{
			name:        "merchant sees masked values",
			role:        "merchant",
			wantTaxID:   "******4321",
			wantAddress: "****",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			taxID := "98-7654321"
			address := "100 Supplier Ave"
			s := &model.Supplier{
				ID:      uuid.New(),
				Name:    "ACME Corp",
				TaxID:   &taxID,
				Address: &address,
			}
			service.MaskSupplier(s, tc.role)

			if *s.TaxID != tc.wantTaxID {
				t.Fatalf("MaskSupplier tax_id = %q, want %q", *s.TaxID, tc.wantTaxID)
			}
			if *s.Address != tc.wantAddress {
				t.Fatalf("MaskSupplier address = %q, want %q", *s.Address, tc.wantAddress)
			}
			// Name should never be masked
			if s.Name != "ACME Corp" {
				t.Fatalf("Name should not be masked: got %q", s.Name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Sensitive data must not leak: verify masked output never contains raw PII
// ---------------------------------------------------------------------------

func TestMaskedOutputNeverContainsRawPII(t *testing.T) {
	rawTaxID := "123-45-6789"
	rawAddress := "742 Evergreen Terrace, Springfield"

	nonAdminRoles := []string{"user", "merchant", "guest"}
	for _, role := range nonAdminRoles {
		t.Run("role="+role, func(t *testing.T) {
			maskedTax := masking.MaskTaxID(rawTaxID, role)
			maskedAddr := masking.MaskAddress(rawAddress, role)

			if strings.Contains(maskedTax, "123-45") {
				t.Fatalf("masked tax_id for %s contains raw prefix: %q", role, maskedTax)
			}
			if strings.Contains(maskedAddr, "Evergreen") {
				t.Fatalf("masked address for %s contains raw content: %q", role, maskedAddr)
			}
		})
	}
}
