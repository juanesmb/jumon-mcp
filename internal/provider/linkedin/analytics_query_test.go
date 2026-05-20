package linkedin

import (
	"strings"
	"testing"
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
