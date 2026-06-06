package meta

import (
	"testing"
)

func TestParseListCreativeAdsRequiresCreativeID(t *testing.T) {
	if _, err := parseListCreativeAdsInput(map[string]any{}); err == nil {
		t.Fatal("expected error for missing creative_id")
	}
	in, err := parseListCreativeAdsInput(map[string]any{"creative_id": "cr_123"})
	if err != nil {
		t.Fatal(err)
	}
	if in.creativeID != "cr_123" {
		t.Fatalf("creative_id = %q", in.creativeID)
	}
	if len(in.fields) != len(defaultCreativeAdListFields) {
		t.Fatalf("got %d default fields", len(in.fields))
	}
}
