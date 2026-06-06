package meta

import (
	"strings"
	"testing"
)

func TestBuildActivitiesQueryTimeRangePrecedence(t *testing.T) {
	in, err := parseActivitiesInput(map[string]any{
		"since": "2026-01-01",
		"until": "2026-01-07",
		"time_range": map[string]any{"since": "2026-02-01", "until": "2026-02-07"},
	})
	if err != nil {
		t.Fatal(err)
	}
	q := buildActivitiesQuery(in)
	if q["time_range"] == "" {
		t.Fatal("expected time_range JSON")
	}
	if _, ok := q["since"]; ok {
		t.Fatal("since should be omitted when time_range is set")
	}
}

func TestBuildActivitiesQuerySinceUntil(t *testing.T) {
	in, err := parseActivitiesInput(map[string]any{
		"since": "2026-03-01",
		"until": "2026-03-31",
	})
	if err != nil {
		t.Fatal(err)
	}
	q := buildActivitiesQuery(in)
	if q["since"] != "2026-03-01" || q["until"] != "2026-03-31" {
		t.Fatalf("since/until = %q / %q", q["since"], q["until"])
	}
	if _, ok := q["time_range"]; ok {
		t.Fatal("time_range should not be set")
	}
}

func TestParseActivitiesDefaultFields(t *testing.T) {
	in, err := parseActivitiesInput(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if len(in.fields) != len(defaultActivityFields) {
		t.Fatalf("got %d fields", len(in.fields))
	}
}

func TestRequireAdSetIDForActivities(t *testing.T) {
	if _, err := requireAdSetID(""); err == nil {
		t.Fatal("expected error for empty adset_id")
	}
	got, err := requireAdSetID(" 789 ")
	if err != nil || got != "789" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestActivitiesQueryIncludesFields(t *testing.T) {
	in, _ := parseActivitiesInput(map[string]any{})
	q := buildActivitiesQuery(in)
	if !strings.Contains(q["fields"], "actor_name") {
		t.Fatalf("fields = %q", q["fields"])
	}
}
