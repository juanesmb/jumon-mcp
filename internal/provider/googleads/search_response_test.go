package googleads

import (
	"strings"
	"testing"
)

func TestIsEmptySearchResults(t *testing.T) {
	if !isEmptySearchResults(map[string]any{"results": []any{}}) {
		t.Fatal("expected empty")
	}
	if isEmptySearchResults(map[string]any{"results": []any{map[string]any{"campaign": map[string]any{"id": "1"}}}}) {
		t.Fatal("expected non-empty")
	}
	if !isEmptySearchResults("not a map") {
		t.Fatal("expected empty for invalid payload")
	}
}

func TestEmptyResultHint_keywords(t *testing.T) {
	hint := emptyResultHint(emptyResultKeywords)
	if !strings.Contains(hint, "keyword_view") {
		t.Fatalf("hint = %q", hint)
	}
	if !strings.Contains(hint, "google_search_pmax_search_terms") {
		t.Fatalf("hint should mention PMax tool: %q", hint)
	}
}

func TestEmptyResultHint_searchTerms(t *testing.T) {
	hint := emptyResultHint(emptyResultSearchTerms)
	if !strings.Contains(hint, "search_term_view") {
		t.Fatalf("hint = %q", hint)
	}
}

func TestCloneSearchRoot_preservesFields(t *testing.T) {
	raw := map[string]any{
		"results":             []any{},
		"totalResultsCount":   "0",
		"fieldMask":           "campaign.id",
	}
	out := cloneSearchRoot(raw)
	if out["totalResultsCount"] != "0" {
		t.Fatalf("lost field: %+v", out)
	}
	if _, ok := out["results"]; !ok {
		t.Fatal("missing results")
	}
}

func TestFormatChannelSummary(t *testing.T) {
	got := formatChannelSummary(map[string]int{
		"SEARCH":       3,
		"DEMAND_GEN":   1,
	})
	if !strings.Contains(got, "SEARCH (3)") || !strings.Contains(got, "DEMAND GEN (1)") {
		t.Fatalf("summary = %q", got)
	}
}
