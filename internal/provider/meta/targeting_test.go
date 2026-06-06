package meta

import (
	"testing"
)

func TestParseInterestListRequiresItems(t *testing.T) {
	if _, err := parseInterestList(map[string]any{}); err == nil {
		t.Fatal("expected error for empty interest_list")
	}
	if _, err := parseInterestList(map[string]any{"interest_list": []any{}}); err == nil {
		t.Fatal("expected error for empty array")
	}
	list, err := parseInterestList(map[string]any{"interest_list": []any{"Basketball", "Sports"}})
	if err != nil || len(list) != 2 {
		t.Fatalf("list = %v err %v", list, err)
	}
}

func TestParseDemographicClassDefault(t *testing.T) {
	if got := parseDemographicClass(map[string]any{}); got != "demographics" {
		t.Fatalf("got %q", got)
	}
	if got := parseDemographicClass(map[string]any{"class": "life_events"}); got != "life_events" {
		t.Fatalf("got %q", got)
	}
}

func TestParseInterestSuggestionsInput(t *testing.T) {
	in, err := parseInterestSuggestionsInput(map[string]any{
		"interest_list": []any{"Fitness"},
		"limit":         10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(in.interestList) != 1 || in.limit != 10 {
		t.Fatalf("got %+v", in)
	}
}

func TestParseSearchLimitDefault(t *testing.T) {
	if got := parseSearchLimit(map[string]any{}); got != defaultListLimit {
		t.Fatalf("got %d", got)
	}
	if got := parseSearchLimit(map[string]any{"limit": 5}); got != 5 {
		t.Fatalf("got %d", got)
	}
}
