package meta

import (
	"testing"
)

func TestParseListCreativeAdsRequiresActAndCreativeID(t *testing.T) {
	if _, err := parseListCreativeAdsInput(map[string]any{}); err == nil {
		t.Fatal("expected error for missing act_id")
	}
	if _, err := parseListCreativeAdsInput(map[string]any{"act_id": "act_1"}); err == nil {
		t.Fatal("expected error for missing creative_id")
	}
	in, err := parseListCreativeAdsInput(map[string]any{
		"act_id":      "act_615183619332316",
		"creative_id": "cr_123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if in.actID != "act_615183619332316" {
		t.Fatalf("act_id = %q", in.actID)
	}
	if in.creativeID != "cr_123" {
		t.Fatalf("creative_id = %q", in.creativeID)
	}
	if len(in.fields) != len(defaultCreativeAdListFields) {
		t.Fatalf("got %d default fields", len(in.fields))
	}
}

func TestFilterAdsByCreativeID(t *testing.T) {
	ads := []any{
		map[string]any{"id": "1", "creative": map[string]any{"id": "cr_a"}},
		map[string]any{"id": "2", "creative": map[string]any{"id": "cr_b"}},
		map[string]any{"id": "3"},
	}
	got := filterAdsByCreativeID(ads, "cr_b")
	if len(got) != 1 {
		t.Fatalf("got %d matches, want 1", len(got))
	}
}
