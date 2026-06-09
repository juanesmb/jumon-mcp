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
