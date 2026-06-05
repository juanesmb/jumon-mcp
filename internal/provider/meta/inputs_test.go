package meta

import (
	"strings"
	"testing"
)

func TestNormalizeActID(t *testing.T) {
	got, err := normalizeActID("12345")
	if err != nil {
		t.Fatal(err)
	}
	if got != "act_12345" {
		t.Fatalf("got %q", got)
	}
	got, err = normalizeActID("act_999")
	if err != nil || got != "act_999" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := normalizeActID(""); err == nil {
		t.Fatal("expected error for empty act_id")
	}
}

func TestParseInsightsInputTimePrecedence(t *testing.T) {
	in, err := parseInsightsInput(map[string]any{
		"date_preset": "last_7d",
		"time_range": map[string]any{"since": "2026-01-01", "until": "2026-01-31"},
	}, defaultInsightsFields)
	if err != nil {
		t.Fatal(err)
	}
	q := buildInsightsQuery(in)
	if _, ok := q["date_preset"]; ok {
		t.Fatal("date_preset should be omitted when time_range set")
	}
	if q["time_range"] == "" {
		t.Fatal("expected time_range JSON")
	}
}

func TestParseAutoPaginateDefaultTrue(t *testing.T) {
	in := parseListPagination(map[string]any{}, defaultCampaignListFields)
	if !in.autoPaginate {
		t.Fatal("expected default auto_paginate true")
	}
}

func TestDefaultInsightsFieldsNoStandaloneActions(t *testing.T) {
	for _, fields := range [][]string{defaultSearchEntitiesFields, defaultInsightsFields} {
		for _, f := range fields {
			if f == "actions" || f == "action_values" {
				t.Fatalf("insights defaults must not include standalone %q", f)
			}
			if strings.Contains(f, ":") {
				t.Fatalf("insights defaults must not include colon action field %q", f)
			}
		}
	}
}

func TestClampInsightsLimit(t *testing.T) {
	if clampInsightsLimit(2000) != maxInsightsLimit {
		t.Fatalf("expected cap at %d", maxInsightsLimit)
	}
	if clampInsightsLimit(50) != 50 {
		t.Fatal("expected 50")
	}
}

func TestParseDeliveryErrorsInput(t *testing.T) {
	in, err := parseDeliveryErrorsInput(map[string]any{"entity_ids": []any{"123", "456"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(in.entityIDs) != 2 {
		t.Fatalf("got %d ids", len(in.entityIDs))
	}
	if _, err := parseDeliveryErrorsInput(map[string]any{}); err == nil {
		t.Fatal("expected error for missing entity_ids")
	}
}
