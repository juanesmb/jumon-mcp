package linkedin

import (
	"context"
	"testing"
)

func TestEnrichCreativesWithLeadForms_annotatesCTAAndFetchesNames(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		pages: []any{
			// batch leadForms?ids=List(42) response
			map[string]any{
				"results": map[string]any{
					"42": map[string]any{
						"id":   "42",
						"name": "Q2 Webinar Signup",
					},
				},
			},
		},
	}

	elements := []any{
		map[string]any{
			"id": "urn:li:sponsoredCreative:1",
			"leadgenCallToAction": map[string]any{
				"label":       "Download",
				"destination": "urn:li:adForm:42",
			},
		},
		map[string]any{
			"id": "urn:li:sponsoredCreative:2",
			// no leadgenCallToAction — should be untouched
		},
	}

	enrichCreativesWithLeadForms(context.Background(), upstream, "user", "tool", elements)

	row0 := elements[0].(map[string]any)
	if row0[fieldLeadFormUrn] != "urn:li:adForm:42" {
		t.Fatalf("leadFormUrn = %v", row0[fieldLeadFormUrn])
	}
	if row0[fieldLeadFormCtaLabel] != "Download" {
		t.Fatalf("leadFormCtaLabel = %v", row0[fieldLeadFormCtaLabel])
	}
	if row0[fieldLeadFormName] != "Q2 Webinar Signup" {
		t.Fatalf("leadFormName = %v", row0[fieldLeadFormName])
	}

	row1 := elements[1].(map[string]any)
	if _, ok := row1[fieldLeadFormUrn]; ok {
		t.Fatal("expected no leadFormUrn on element without CTA")
	}

	if upstream.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1 batch leadForms call", upstream.calls)
	}
	if upstream.paths[0] != "leadForms" {
		t.Fatalf("path = %q, want \"leadForms\"", upstream.paths[0])
	}
}

func TestEnrichCreativesWithLeadForms_deduplicatesBatchIDs(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"results": map[string]any{
					"7": map[string]any{"name": "Form Seven"},
				},
			},
		},
	}

	// Two creatives share the same form ID.
	elements := []any{
		map[string]any{
			"leadgenCallToAction": map[string]any{"destination": "urn:li:adForm:7"},
		},
		map[string]any{
			"leadgenCallToAction": map[string]any{"destination": "urn:li:adForm:7"},
		},
	}

	enrichCreativesWithLeadForms(context.Background(), upstream, "user", "tool", elements)

	if upstream.calls != 1 {
		t.Fatalf("upstream calls = %d, want exactly 1 (deduped)", upstream.calls)
	}
	for i, item := range elements {
		row := item.(map[string]any)
		if row[fieldLeadFormName] != "Form Seven" {
			t.Fatalf("element[%d] leadFormName = %v", i, row[fieldLeadFormName])
		}
	}
}

func TestEnrichCreativesWithLeadForms_skipsBatchWhenNoAdFormURNs(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{}

	// Versioned form URN cannot be resolved via batch GET.
	elements := []any{
		map[string]any{
			"leadgenCallToAction": map[string]any{
				"destination": "urn:li:versionedLeadGenForm:(urn:li:leadGenForm:3162,1)",
			},
		},
	}

	enrichCreativesWithLeadForms(context.Background(), upstream, "user", "tool", elements)

	row := elements[0].(map[string]any)
	if _, ok := row[fieldLeadFormName]; ok {
		t.Fatal("expected no leadFormName for versioned URN")
	}
	// formURN should still be set
	if row[fieldLeadFormUrn] == nil {
		t.Fatal("expected leadFormUrn to be set even for versioned URN")
	}
	if upstream.calls != 0 {
		t.Fatalf("upstream calls = %d, want 0 (no batchable IDs)", upstream.calls)
	}
}

func TestEnrichCreativesWithLeadForms_batchFailureDoesNotBreakParent(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		// Return empty results map — no names resolved.
		pages: []any{map[string]any{"results": map[string]any{}}},
	}

	elements := []any{
		map[string]any{
			"leadgenCallToAction": map[string]any{
				"label":       "Sign Up",
				"destination": "urn:li:adForm:99",
			},
		},
	}

	enrichCreativesWithLeadForms(context.Background(), upstream, "user", "tool", elements)

	row := elements[0].(map[string]any)
	if row[fieldLeadFormUrn] != "urn:li:adForm:99" {
		t.Fatalf("leadFormUrn = %v", row[fieldLeadFormUrn])
	}
	if _, ok := row[fieldLeadFormName]; ok {
		t.Fatal("expected no leadFormName when batch returns no results")
	}
}

func TestEnrichCreativesResponse_includesLeadFormDetailsWhenEnabled(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		pages: []any{
			// First call is adPreviews (previewURL), second is leadForms batch.
			map[string]any{
				"elements": []any{
					map[string]any{
						"preview": "<iframe src='https://www.linkedin.com/ads/ad-preview/?d=x'></iframe>",
						"placement": map[string]any{
							"linkedin": map[string]any{
								"placementName":           "FEED",
								"contentPresentationType": "DESKTOP_WEBSITE",
							},
						},
					},
				},
			},
			map[string]any{
				"results": map[string]any{
					"55": map[string]any{"name": "Test Form"},
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
				"leadgenCallToAction": map[string]any{
					"label":       "Learn More",
					"destination": "urn:li:adForm:55",
				},
			},
		},
	}

	enriched := enrichCreativesResponse(
		context.Background(), upstream, "user", "tool", "512247261",
		payload,
		enrichCreativesOptions{
			includePreviewURLs:     true,
			includeLeadFormDetails: true,
		},
	)

	row := enriched.(map[string]any)["elements"].([]any)[0].(map[string]any)
	if row[fieldLeadFormUrn] != "urn:li:adForm:55" {
		t.Fatalf("leadFormUrn = %v", row[fieldLeadFormUrn])
	}
	if row[fieldLeadFormName] != "Test Form" {
		t.Fatalf("leadFormName = %v", row[fieldLeadFormName])
	}
	if row[fieldLeadFormCtaLabel] != "Learn More" {
		t.Fatalf("leadFormCtaLabel = %v", row[fieldLeadFormCtaLabel])
	}
	if upstream.calls != 2 {
		t.Fatalf("upstream calls = %d, want 2 (1 preview + 1 leadForms batch)", upstream.calls)
	}
}

func TestParseLeadFormBatchResponse_elementsListFallback(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"elements": []any{
			map[string]any{"id": "12", "name": "Form Twelve"},
		},
	}
	result := parseLeadFormBatchResponse(raw)
	if result["12"].name != "Form Twelve" {
		t.Fatalf("form name = %q", result["12"].name)
	}
}

func TestParseLeadFormBatchResponse_multiLocaleString(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"results": map[string]any{
			"5": map[string]any{
				"id": "5",
				"name": map[string]any{
					"localized": map[string]any{
						"en_US": "Localized Form",
					},
				},
			},
		},
	}
	result := parseLeadFormBatchResponse(raw)
	if result["5"].name != "Localized Form" {
		t.Fatalf("form name = %q", result["5"].name)
	}
}
