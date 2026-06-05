package meta

import (
	"strings"
	"testing"
)

func TestBuildInsightsQueryCSVAndJSON(t *testing.T) {
	in := insightsInput{
		fields:     []string{"impressions", "spend"},
		datePreset: "last_30d",
		breakdowns: []string{"publisher_platform"},
		filtering: []map[string]any{
			{"field": "impressions", "operator": "GREATER_THAN", "value": 0},
		},
		limit: 50,
	}
	q := buildInsightsQuery(in)
	if q["fields"] != "impressions,spend" {
		t.Fatalf("fields = %q", q["fields"])
	}
	if q["breakdowns"] != "publisher_platform" {
		t.Fatalf("breakdowns = %q", q["breakdowns"])
	}
	if !strings.Contains(q["filtering"], "GREATER_THAN") {
		t.Fatalf("filtering = %q", q["filtering"])
	}
	if q["date_preset"] != "last_30d" {
		t.Fatalf("date_preset = %q", q["date_preset"])
	}
}

func TestBuildListQueryEffectiveStatus(t *testing.T) {
	in := listPaginationInput{
		fields:          []string{"id", "name"},
		limit:           25,
		effectiveStatus: []string{"ACTIVE", "PAUSED"},
	}
	q := buildListQuery(in)
	if !strings.Contains(q["effective_status"], "ACTIVE") {
		t.Fatalf("effective_status = %q", q["effective_status"])
	}
}
