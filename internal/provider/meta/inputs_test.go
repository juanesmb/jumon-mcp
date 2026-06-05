package meta

import "testing"

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
