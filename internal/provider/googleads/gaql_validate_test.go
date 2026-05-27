package googleads

import (
	"strings"
	"testing"
)

func TestNormalizeGAQLResourceName(t *testing.T) {
	name, err := normalizeGAQLResourceName("campaign")
	if err != nil || name != "campaign" {
		t.Fatalf("expected campaign, got %q err=%v", name, err)
	}
	if _, err := normalizeGAQLResourceName("not_a_real_resource"); err == nil {
		t.Fatal("expected error for unknown resource")
	}
	if _, err := normalizeGAQLResourceName("campaign; DROP"); err == nil {
		t.Fatal("expected error for invalid characters")
	}
}

func TestValidateGAQLFieldName(t *testing.T) {
	if err := validateGAQLFieldName("campaign.id", "campaign"); err != nil {
		t.Fatalf("valid field rejected: %v", err)
	}
	if err := validateGAQLFieldName("metrics.clicks", "campaign"); err != nil {
		t.Fatalf("metrics field rejected: %v", err)
	}
	if err := validateGAQLFieldName("id", "campaign"); err == nil {
		t.Fatal("bare id should fail")
	}
	if err := validateGAQLFieldName("campaign.*", "campaign"); err == nil {
		t.Fatal("wildcard should fail")
	}
}

func TestBuildGenericSearchQuery(t *testing.T) {
	in := gaqlSearchInput{
		resource:   "keyword_view",
		fields:     []string{"ad_group_criterion.keyword.text", "metrics.clicks"},
		conditions: []string{"campaign.id = 123", "segments.date BETWEEN '2026-01-01' AND '2026-01-31'"},
		orderings:  []string{"metrics.clicks DESC"},
		limit:      100,
	}
	query := buildGenericSearchQuery(in, "keyword_view")
	if !strings.Contains(query, "FROM keyword_view") {
		t.Fatalf("missing FROM: %s", query)
	}
	if !strings.Contains(query, "WHERE campaign.id = 123 AND segments.date") {
		t.Fatalf("missing WHERE: %s", query)
	}
	if !strings.Contains(query, "ORDER BY metrics.clicks DESC") {
		t.Fatalf("missing ORDER BY: %s", query)
	}
	if !strings.Contains(query, "LIMIT 100") {
		t.Fatalf("missing LIMIT: %s", query)
	}
	if !strings.Contains(query, "PARAMETERS omit_unselected_resource_names=true") {
		t.Fatalf("missing PARAMETERS: %s", query)
	}
}

func TestNormalizeGenericSearchLimitChangeEvent(t *testing.T) {
	if got := normalizeGenericSearchLimit("change_event", 20000); got != maxChangeEventLimit {
		t.Fatalf("change_event cap = %d, want %d", got, maxChangeEventLimit)
	}
}

func TestMetricsDateHint(t *testing.T) {
	in := gaqlSearchInput{
		fields:     []string{"metrics.clicks"},
		conditions: []string{"campaign.status = ENABLED"},
	}
	if hint := metricsDateHint(in); hint == "" {
		t.Fatal("expected date hint")
	}
	in.conditions = append(in.conditions, "segments.date BETWEEN '2026-01-01' AND '2026-01-31'")
	if hint := metricsDateHint(in); hint != "" {
		t.Fatalf("unexpected hint: %s", hint)
	}
}

func TestGAQLResourceAllowlistLoaded(t *testing.T) {
	if len(gaqlResourceAllowlist) < 100 {
		t.Fatalf("allowlist too small: %d", len(gaqlResourceAllowlist))
	}
	if !isAllowedGAQLResource("search_term_view") {
		t.Fatal("search_term_view should be allowed")
	}
}

func TestValidateGAQLSearchInput(t *testing.T) {
	in := gaqlSearchInput{
		resource: "campaign",
		fields:   []string{"campaign.id", "metrics.impressions"},
		limit:    500,
	}
	resource, err := validateGAQLSearchInput(in)
	if err != nil || resource != "campaign" {
		t.Fatalf("validate failed: resource=%q err=%v", resource, err)
	}
}

func TestValidateGAQLSearchInputRejectsBadResource(t *testing.T) {
	in := gaqlSearchInput{
		resource: "invalid_resource_xyz",
		fields:   []string{"campaign.id"},
	}
	if _, err := validateGAQLSearchInput(in); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPathGoogleAdsFieldsSearch(t *testing.T) {
	got := pathGoogleAdsFieldsSearch("v22")
	if got != "v22/googleAdsFields:search" {
		t.Fatalf("path = %q", got)
	}
}
