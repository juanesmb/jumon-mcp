package meta

import (
	"fmt"
	"testing"
)

func TestBuildDeliveryErrorEntry(t *testing.T) {
	entry := buildDeliveryErrorEntry("123", map[string]any{
		"id":                     "456",
		"name":                   "Ad 1",
		"failed_delivery_checks": []any{map[string]any{"summary": "blocked"}},
	})
	if entry["entity_id"] != "456" {
		t.Fatalf("entity_id = %v", entry["entity_id"])
	}
	if entry["entity_type"] != "ad" {
		t.Fatalf("entity_type = %v", entry["entity_type"])
	}
}

func TestInferEntityType(t *testing.T) {
	if got := inferEntityType(map[string]any{"failed_delivery_checks": []any{}}); got != "ad" {
		t.Fatalf("got %q", got)
	}
	if got := inferEntityType(map[string]any{"issues_info": []any{}}); got != "campaign_or_adset" {
		t.Fatalf("got %q", got)
	}
	if got := inferEntityType(map[string]any{"name": "x"}); got != "unknown" {
		t.Fatalf("got %q", got)
	}
}

func TestIsGraphNonexistingFieldError(t *testing.T) {
	if !isGraphNonexistingFieldError(fmt.Errorf("meta api returned status 400: (#100) Tried accessing nonexisting field (failed_delivery_checks)")) {
		t.Fatal("expected nonexisting field detection")
	}
	if isGraphNonexistingFieldError(fmt.Errorf("meta api returned status 429")) {
		t.Fatal("rate limit should not match")
	}
}
