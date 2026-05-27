package googleads

import "testing"

func TestGoogleNormalizeCustomerID(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"customers/123-456-7890", "1234567890"},
		{"1234567890", "1234567890"},
		{"  customers/999  ", "999"},
	}
	for _, tc := range tests {
		if got := googleNormalizeCustomerID(tc.in); got != tc.want {
			t.Fatalf("normalize(%q) = %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseCustomerContextRequiresID(t *testing.T) {
	_, err := parseCustomerContext(map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseResolveCustomerInput(t *testing.T) {
	in, err := parseResolveCustomerInput(map[string]any{
		"account_name": "Acme",
		"match_mode":   "exact",
	})
	if err != nil {
		t.Fatal(err)
	}
	if in.accountName != "Acme" || in.matchMode != "exact" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestParseReportFiltersLimit(t *testing.T) {
	in, err := parseReportFilters(map[string]any{
		"customer_id": "123",
		"limit":       float64(100),
	})
	if err != nil {
		t.Fatal(err)
	}
	if in.limit != 100 {
		t.Fatalf("limit = %d", in.limit)
	}
}

func TestParseConversionPerformanceDefaultsDateRange(t *testing.T) {
	in, err := parseConversionPerformanceFilters(map[string]any{"customer_id": "123"})
	if err != nil {
		t.Fatal(err)
	}
	if in.dateRangeStart == "" || in.dateRangeEnd == "" {
		t.Fatal("expected default date range")
	}
}

func TestMatchAccountName(t *testing.T) {
	if !matchAccountName("Acme US", "acme", "contains") {
		t.Fatal("expected contains match")
	}
	if matchAccountName("Acme US", "Acme UK", "exact") {
		t.Fatal("expected no exact match")
	}
	if matchAccountName("Acme US", "Acme US", "exact") {
		// ok
	} else {
		t.Fatal("expected exact match")
	}
}
