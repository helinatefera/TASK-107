package masking

import "strings"

func MaskTaxID(taxID string, role string) string {
	if taxID == "" {
		return ""
	}
	if role == "administrator" {
		return taxID
	}
	if len(taxID) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(taxID)-4) + taxID[len(taxID)-4:]
}

func MaskAddress(addr string, role string) string {
	if addr == "" {
		return ""
	}
	if role == "administrator" {
		return addr
	}
	return "****"
}
