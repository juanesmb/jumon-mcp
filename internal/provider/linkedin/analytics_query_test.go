package linkedin

import (
	"strings"
	"testing"
	"time"
)

func TestApplyAnalyticsPivots_statisticsMultiPivot(t *testing.T) {
	t.Parallel()

	query := map[string]string{}
	if err := applyAnalyticsPivots(query, finderTypeStatistics, []string{"CAMPAIGN", "PLACEMENT_NAME"}); err != nil {
		t.Fatalf("applyAnalyticsPivots() error = %v", err)
	}
	if got := query["pivots"]; got != "List(CAMPAIGN,PLACEMENT_NAME)" {
		t.Fatalf("pivots = %q, want List(CAMPAIGN,PLACEMENT_NAME)", got)
	}
	if _, ok := query["pivot"]; ok {
		t.Fatal("expected no pivot key for statistics finder")
	}
}

func TestApplyAnalyticsPivots_revenueMultiPivot(t *testing.T) {
	t.Parallel()

	query := map[string]string{}
	if err := applyAnalyticsPivots(query, finderTypeAttributedRevenue, []string{"CAMPAIGN", "CAMPAIGN_GROUP"}); err != nil {
		t.Fatalf("applyAnalyticsPivots() error = %v", err)
	}
	if got := query["pivots"]; got != "List(CAMPAIGN,CAMPAIGN_GROUP)" {
		t.Fatalf("pivots = %q", got)
	}
}

func TestApplyAnalyticsPivots_revenueRejectsInvalidPivot(t *testing.T) {
	t.Parallel()

	err := applyAnalyticsPivots(map[string]string{}, finderTypeAttributedRevenue, []string{"CREATIVE"})
	if err == nil {
		t.Fatal("expected error for invalid revenue pivot")
	}
}

func TestApplyAnalyticsPivots_analyticsUsesFirstPivotOnly(t *testing.T) {
	t.Parallel()

	query := map[string]string{}
	if err := applyAnalyticsPivots(query, finderTypeAnalytics, []string{"CAMPAIGN", "PLACEMENT_NAME"}); err != nil {
		t.Fatalf("applyAnalyticsPivots() error = %v", err)
	}
	if got := query["pivot"]; got != "CAMPAIGN" {
		t.Fatalf("pivot = %q, want CAMPAIGN", got)
	}
	if _, ok := query["pivots"]; ok {
		t.Fatal("expected no pivots key for analytics finder")
	}
}

func TestApplyAnalyticsPivots_statisticsPivotCountValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pivots []string
	}{
		{name: "zero pivots", pivots: nil},
		{name: "four pivots", pivots: []string{"A", "B", "C", "D"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := applyAnalyticsPivots(map[string]string{}, finderTypeStatistics, tt.pivots)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestBuildAdAnalyticsQuery_statisticsIncludesFacets(t *testing.T) {
	t.Parallel()

	query, err := buildAdAnalyticsQuery(getAnalyticsInput{
		accountID:  "12345",
		startDate:  "2026-04-27",
		endDate:    "2026-05-03",
		finderType: finderTypeStatistics,
		pivots:     []string{"CAMPAIGN", "PLACEMENT_NAME"},
		fields:     []string{"impressions", "clicks"},
	})
	if err != nil {
		t.Fatalf("buildAdAnalyticsQuery() error = %v", err)
	}

	if query["q"] != finderTypeStatistics {
		t.Fatalf("q = %q, want statistics", query["q"])
	}
	if query["pivots"] != "List(CAMPAIGN,PLACEMENT_NAME)" {
		t.Fatalf("pivots = %q", query["pivots"])
	}
	if !strings.Contains(query["accounts"], "12345") {
		t.Fatalf("accounts = %q", query["accounts"])
	}
	if !strings.Contains(query["dateRange"], "year:2026") {
		t.Fatalf("dateRange = %q", query["dateRange"])
	}
	if !strings.Contains(query["fields"], "pivotValues") {
		t.Fatalf("fields = %q, expected pivotValues auto-included", query["fields"])
	}
}

func TestBuildAdAnalyticsQuery_revenueFinderUsesAccountAndRevenueFields(t *testing.T) {
	t.Parallel()

	query, err := buildAdAnalyticsQuery(getAnalyticsInput{
		accountID:  "12345",
		startDate:  "2025-01-01",
		endDate:    "2025-04-01",
		finderType: finderTypeAttributedRevenue,
		pivots:     []string{"CAMPAIGN"},
	})
	if err != nil {
		t.Fatalf("buildAdAnalyticsQuery() error = %v", err)
	}

	if query["q"] != finderTypeAttributedRevenue {
		t.Fatalf("q = %q", query["q"])
	}
	if !strings.Contains(query["account"], "12345") {
		t.Fatalf("account = %q, want account facet", query["account"])
	}
	if _, ok := query["accounts"]; ok {
		t.Fatal("expected no accounts key for revenue finder")
	}
	if query["pivots"] != "List(CAMPAIGN)" {
		t.Fatalf("pivots = %q", query["pivots"])
	}
	if !strings.Contains(query["fields"], "revenueAttributionMetrics") {
		t.Fatalf("fields = %q", query["fields"])
	}
	if strings.Contains(query["fields"], "approximateMemberReach") {
		t.Fatalf("fields should not include reach: %q", query["fields"])
	}
	if _, ok := query["sortBy"]; ok {
		t.Fatal("sortBy should be omitted for revenue finder")
	}
	if _, ok := query["timeGranularity"]; ok {
		t.Fatal("timeGranularity should be omitted for revenue finder")
	}
}

func TestBuildAdAnalyticsQuery_revenueFinderNormalizesNestedFieldNames(t *testing.T) {
	t.Parallel()

	query, err := buildAdAnalyticsQuery(getAnalyticsInput{
		accountID:       "12345",
		startDate:       "2025-01-01",
		endDate:         "2025-04-01",
		finderType:      finderTypeAttributedRevenue,
		pivots:          []string{"CAMPAIGN"},
		fields:          []string{"returnOnAdSpend", "revenueWonInUsd"},
		timeGranularity: "ALL",
	})
	if err != nil {
		t.Fatalf("buildAdAnalyticsQuery() error = %v", err)
	}

	want := "revenueAttributionMetrics:(returnOnAdSpend,revenueWonInUsd)"
	if !strings.Contains(query["fields"], want) {
		t.Fatalf("fields = %q, want nested projection %q", query["fields"], want)
	}
}

func TestBuildAdAnalyticsQuery_revenueDateRangeTooShort(t *testing.T) {
	t.Parallel()

	_, err := buildAdAnalyticsQuery(getAnalyticsInput{
		accountID:  "12345",
		startDate:  "2026-04-27",
		endDate:    "2026-05-03",
		finderType: finderTypeAttributedRevenue,
		pivots:     []string{"CAMPAIGN"},
	})
	if err == nil {
		t.Fatal("expected error for short revenue date range")
	}
}

func TestBuildAdAnalyticsQuery_demographicPivotSkipsReachDefaults(t *testing.T) {
	t.Parallel()

	query, err := buildAdAnalyticsQuery(getAnalyticsInput{
		accountID:  "12345",
		startDate:  "2026-04-27",
		endDate:    "2026-05-03",
		finderType: finderTypeAnalytics,
		pivots:     []string{"MEMBER_INDUSTRY"},
	})
	if err != nil {
		t.Fatalf("buildAdAnalyticsQuery() error = %v", err)
	}

	if strings.Contains(query["fields"], "approximateMemberReach") {
		t.Fatalf("fields should not include reach for demographic pivot: %q", query["fields"])
	}
	if !strings.Contains(query["fields"], "impressions") {
		t.Fatalf("fields = %q", query["fields"])
	}
}

func TestBuildAdAnalyticsQuery_missingAccountID(t *testing.T) {
	t.Parallel()

	_, err := buildAdAnalyticsQuery(getAnalyticsInput{
		startDate: "2026-04-27",
		pivots:    []string{"CAMPAIGN"},
	})
	if err == nil {
		t.Fatal("expected error for missing account_id")
	}
}

func TestValidateRevenueDateRange(t *testing.T) {
	t.Parallel()

	start := mustParseTestDate(t, "2025-01-01")
	end := mustParseTestDate(t, "2025-01-30")
	if err := validateRevenueDateRange(start, end); err != nil {
		t.Fatalf("30-day range should pass: %v", err)
	}

	shortEnd := mustParseTestDate(t, "2025-01-20")
	if err := validateRevenueDateRange(start, shortEnd); err == nil {
		t.Fatal("expected error for 20-day range")
	}
}

func mustParseTestDate(t *testing.T, raw string) time.Time {
	t.Helper()
	d, err := parseDate(raw)
	if err != nil {
		t.Fatalf("parseDate(%q): %v", raw, err)
	}
	return d
}
