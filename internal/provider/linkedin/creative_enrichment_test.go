package linkedin

import (
	"context"
	"testing"
)

func TestEnrichCreativesResponse_addsFeedAndPreviewURLs(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"elements": []any{
					map[string]any{
						"preview": "<iframe src='https://www.linkedin.com/ads/ad-preview/?d=abc'></iframe>",
						"placement": map[string]any{
							"linkedin": map[string]any{
								"placementName":           "FEED",
								"contentPresentationType": "DESKTOP_WEBSITE",
							},
						},
					},
				},
			},
		},
	}

	payload := map[string]any{
		"elements": []any{
			map[string]any{
				"id": "urn:li:sponsoredCreative:123",
				"content": map[string]any{
					"reference": "urn:li:ugcPost:789",
				},
			},
		},
	}
	enriched := enrichCreativesResponse(
		context.Background(),
		upstream,
		"user",
		"linkedin_search_creatives",
		"512247261",
		payload,
		enrichCreativesOptions{includePreviewURLs: true},
	)

	row := enriched.(map[string]any)["elements"].([]any)[0].(map[string]any)
	if row[fieldFeedURL] != "https://www.linkedin.com/feed/update/urn:li:ugcPost:789" {
		t.Fatalf("feedUrl = %v", row[fieldFeedURL])
	}
	if row[fieldPreviewURL] != "https://www.linkedin.com/ads/ad-preview/?d=abc" {
		t.Fatalf("previewUrl = %v", row[fieldPreviewURL])
	}
	if upstream.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1", upstream.calls)
	}
	if upstream.paths[0] != "adPreviews" {
		t.Fatalf("preview call path = %q", upstream.paths[0])
	}
}

func TestEnrichCreativesResponse_skipsPreviewWhenDisabled(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{}

	payload := map[string]any{
		"elements": []any{
			map[string]any{
				"id": "urn:li:sponsoredCreative:123",
				"content": map[string]any{
					"reference": "urn:li:share:456",
				},
			},
		},
	}

	enriched := enrichCreativesResponse(
		context.Background(),
		upstream,
		"user",
		"tool",
		"512247261",
		payload,
		enrichCreativesOptions{includePreviewURLs: false},
	)

	row := enriched.(map[string]any)["elements"].([]any)[0].(map[string]any)
	if _, ok := row[fieldPreviewURL]; ok {
		t.Fatal("expected previewUrl to be omitted")
	}
	if upstream.calls != 0 {
		t.Fatalf("upstream calls = %d, want 0 preview calls", upstream.calls)
	}
}

func TestEnrichCreativesResponse_previewFailureDoesNotBreakSearch(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{"elements": []any{}},
		},
	}

	payload := map[string]any{
		"elements": []any{
			map[string]any{
				"id": "urn:li:sponsoredCreative:123",
			},
		},
	}

	enriched := enrichCreativesResponse(
		context.Background(),
		upstream,
		"user",
		"tool",
		"512247261",
		payload,
		enrichCreativesOptions{includePreviewURLs: true},
	)

	row := enriched.(map[string]any)["elements"].([]any)[0].(map[string]any)
	if _, ok := row[fieldPreviewURL]; ok {
		t.Fatal("expected previewUrl to be omitted on preview failure")
	}
	if upstream.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1 preview call on failure", upstream.calls)
	}
}
