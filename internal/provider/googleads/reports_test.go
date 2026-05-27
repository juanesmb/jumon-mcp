package googleads

import (
	"strings"
	"testing"
)

func TestSearchKeywordsQuery(t *testing.T) {
	query := buildKeywordsQuery(reportFilters{
		customerContext: customerContext{customerID: "123"},
		keywordContains: "shoes",
		limit:           100,
		dateRangeStart:  "2026-01-01",
		dateRangeEnd:    "2026-01-31",
	})
	if !strings.Contains(query, "FROM keyword_view") {
		t.Fatalf("missing resource: %s", query)
	}
	if !strings.Contains(query, "segments.date BETWEEN") {
		t.Fatalf("missing date segment: %s", query)
	}
	if !strings.Contains(query, "LIMIT 100") {
		t.Fatalf("missing limit: %s", query)
	}
}

func TestSearchSearchTermsQuery(t *testing.T) {
	query := buildSearchTermsQuery(reportFilters{
		customerContext:    customerContext{customerID: "123"},
		searchTermContains: "running",
		limit:              200,
	})
	if !strings.Contains(query, "FROM search_term_view") {
		t.Fatalf("missing resource: %s", query)
	}
	if !strings.Contains(query, "search_term_view.search_term LIKE '%running%'") {
		t.Fatalf("missing filter: %s", query)
	}
}

func TestListConversionActionsQuery(t *testing.T) {
	query := buildConversionActionsQuery(reportFilters{
		customerContext: customerContext{customerID: "123"},
		nameContains:    "Purchase",
		limit:           50,
	})
	if !strings.Contains(query, "FROM conversion_action") {
		t.Fatalf("missing resource: %s", query)
	}
	if strings.Contains(query, "segments.date") {
		t.Fatal("conversion_action catalog should not include date segments")
	}
}

func TestSearchConversionPerformanceQuery(t *testing.T) {
	query := buildConversionPerformanceQuery(reportFilters{
		customerContext:     customerContext{customerID: "123"},
		conversionActionIDs: []string{"99"},
		dateRangeStart:      "2026-01-01",
		dateRangeEnd:        "2026-01-31",
		limit:               300,
	})
	if !strings.Contains(query, "segments.conversion_action IN ('customers/123/conversionActions/99')") {
		t.Fatalf("missing conversion action filter: %s", query)
	}
	if !strings.Contains(query, "metrics.conversions") {
		t.Fatalf("missing metrics: %s", query)
	}
}
