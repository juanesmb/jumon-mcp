package linkedin

import (
	"context"
	"testing"
)

func TestParseAnalyticsStartToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		token string
		want  string
		ok    bool
	}{
		{token: "10", want: "10", ok: true},
		{token: "start=10", want: "10", ok: true},
		{
			token: "/rest/adAnalytics?q=analytics&start=20&count=10&pivot=CAMPAIGN",
			want:  "20",
			ok:    true,
		},
		{token: "", ok: false},
		{token: "not-a-number", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			got, ok := parseAnalyticsStartToken(tt.token)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("start = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextAnalyticsStart_fromPagingLink(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"paging": map[string]any{
			"links": []any{
				map[string]any{
					"rel":  "next",
					"href": "/rest/adAnalytics?q=analytics&start=10&count=10&pivot=CAMPAIGN",
				},
			},
		},
	}
	if got := nextAnalyticsStart(raw); got != "10" {
		t.Fatalf("next start = %q", got)
	}
}

func TestFetchAnalyticsPages_autoPaginateMergesElements(t *testing.T) {
	t.Parallel()

	stub := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"elements": []any{map[string]any{"pivotValues": []any{"urn:li:sponsoredCampaign:1"}}},
				"paging": map[string]any{
					"start": 0,
					"count": 10,
					"links": []any{
						map[string]any{
							"rel":  "next",
							"href": "/rest/adAnalytics?start=10&count=10",
						},
					},
				},
			},
			map[string]any{
				"elements": []any{map[string]any{"pivotValues": []any{"urn:li:sponsoredCampaign:2"}}},
				"paging": map[string]any{
					"start": 10,
					"count": 10,
					"links": []any{},
				},
			},
		},
	}

	query := map[string]string{"q": "analytics", "count": "10"}
	result, err := fetchAnalyticsPages(context.Background(), stub, "user", "tool", query, true)
	if err != nil {
		t.Fatalf("fetchAnalyticsPages() error = %v", err)
	}

	page, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	elements, ok := page["elements"].([]any)
	if !ok || len(elements) != 2 {
		t.Fatalf("elements = %#v, want 2 items", page["elements"])
	}
	if stub.calls != 2 {
		t.Fatalf("request calls = %d, want 2", stub.calls)
	}
	if stub.queries[1]["start"] != "10" {
		t.Fatalf("second query start = %q, want 10", stub.queries[1]["start"])
	}

	paging, ok := page["paging"].(map[string]any)
	if !ok {
		t.Fatal("expected paging")
	}
	links, ok := paging["links"].([]any)
	if !ok || len(links) != 0 {
		t.Fatalf("links = %#v, want next link stripped", paging["links"])
	}
}

func TestFetchAnalyticsPages_stopsWhenPageRepeatsPivots(t *testing.T) {
	t.Parallel()

	stub := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"elements": []any{map[string]any{"pivotValues": []any{"urn:li:sponsoredCampaign:1"}}},
				"paging": map[string]any{
					"links": []any{
						map[string]any{"rel": "next", "href": "/rest/adAnalytics?start=10&count=10"},
					},
				},
			},
			map[string]any{
				"elements": []any{map[string]any{"pivotValues": []any{"urn:li:sponsoredCampaign:1"}}},
				"paging": map[string]any{
					"links": []any{
						map[string]any{"rel": "next", "href": "/rest/adAnalytics?start=20&count=10"},
					},
				},
			},
		},
	}

	query := map[string]string{"q": "analytics", "count": "10"}
	result, err := fetchAnalyticsPages(context.Background(), stub, "user", "tool", query, true)
	if err != nil {
		t.Fatalf("fetchAnalyticsPages() error = %v", err)
	}

	page, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	elements, ok := page["elements"].([]any)
	if !ok || len(elements) != 1 {
		t.Fatalf("elements = %#v, want 1 deduplicated item", page["elements"])
	}
	if stub.calls != 2 {
		t.Fatalf("request calls = %d, want 2", stub.calls)
	}
}

func TestApplyAnalyticsPagination_setsStartAndCount(t *testing.T) {
	t.Parallel()

	query := map[string]string{"q": "analytics"}
	applyAnalyticsPagination(query, "10", 10)
	if query["start"] != "10" {
		t.Fatalf("start = %q, want 10", query["start"])
	}
	if query["count"] != "10" {
		t.Fatalf("count = %q, want 10", query["count"])
	}
}

func TestGetAnalytics_autoPaginateMergesPages(t *testing.T) {
	t.Parallel()

	stub := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"elements": []any{map[string]any{"impressions": float64(1)}},
				"paging": map[string]any{
					"links": []any{
						map[string]any{"rel": "next", "href": "/rest/adAnalytics?start=10&count=10"},
					},
				},
			},
			map[string]any{
				"elements": []any{map[string]any{"impressions": float64(2)}},
				"paging":   map[string]any{"links": []any{}},
			},
		},
	}

	svc := &service{proxy: stub}
	result, err := svc.getAnalytics(context.Background(), "user", "linkedin_get_ad_analytics", getAnalyticsInput{
		accountID:       "512247261",
		startDate:       "2025-01-01",
		endDate:         "2025-05-26",
		pivots:          []string{"CAMPAIGN"},
		autoPaginate:    true,
		pageSize:        10,
		timeGranularity: "ALL",
	})
	if err != nil {
		t.Fatalf("getAnalytics() error = %v", err)
	}

	page, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	elements, ok := page["elements"].([]any)
	if !ok || len(elements) != 2 {
		t.Fatalf("elements = %#v, want 2 merged rows", page["elements"])
	}
	if stub.queries[0]["count"] != "10" {
		t.Fatalf("count = %q", stub.queries[0]["count"])
	}
}
