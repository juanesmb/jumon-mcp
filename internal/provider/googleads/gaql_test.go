package googleads

import "testing"

func TestGoogleBuildWhereClause(t *testing.T) {
	got := googleBuildWhereClause([]string{"", "campaign.status IN (ENABLED)", "segments.date >= '2026-01-01'"})
	want := "WHERE campaign.status IN (ENABLED) AND segments.date >= '2026-01-01'"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if googleBuildWhereClause(nil) != "" {
		t.Fatal("expected empty where clause")
	}
}

func TestGoogleInClause(t *testing.T) {
	if googleInClause("campaign.id", nil) != "" {
		t.Fatal("expected empty")
	}
	got := googleInClause("campaign.id", []string{"1", "2"})
	if got != "campaign.id IN (1,2)" {
		t.Fatalf("got %q", got)
	}
}

func TestGoogleQuotedInClause(t *testing.T) {
	got := googleQuotedInClause("segments.conversion_action", []string{"customers/1/conversionActions/2"})
	want := "segments.conversion_action IN ('customers/1/conversionActions/2')"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGoogleLikeClause(t *testing.T) {
	got := googleLikeClause("campaign.name", "Summer")
	if got != "campaign.name LIKE '%Summer%'" {
		t.Fatalf("got %q", got)
	}
}

func TestGoogleExactClause(t *testing.T) {
	got := googleExactClause("customer_client.descriptive_name", "Acme")
	if got != "customer_client.descriptive_name = 'Acme'" {
		t.Fatalf("got %q", got)
	}
}

func TestGoogleDateBetweenClause(t *testing.T) {
	got := googleDateBetweenClause("2026-01-01", "2026-01-31")
	if got != "segments.date BETWEEN '2026-01-01' AND '2026-01-31'" {
		t.Fatalf("got %q", got)
	}
}

func TestGoogleConversionActionResourceInClause(t *testing.T) {
	got := googleConversionActionResourceInClause("123", []string{"456"})
	want := "segments.conversion_action IN ('customers/123/conversionActions/456')"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeReportLimit(t *testing.T) {
	if normalizeReportLimit(0) != defaultReportLimit {
		t.Fatal("expected default")
	}
	if normalizeReportLimit(5000) != maxReportLimit {
		t.Fatal("expected max cap")
	}
}
