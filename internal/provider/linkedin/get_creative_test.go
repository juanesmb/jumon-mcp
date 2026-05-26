package linkedin

import (
	"context"
	"testing"
)

func TestParseCreativeBatchGetResponse_extractsByURN(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"results": map[string]any{
			"urn:li:sponsoredCreative:586489603": map[string]any{
				"id":   "urn:li:sponsoredCreative:586489603",
				"name": "Ad",
			},
		},
	}
	row, ok := parseCreativeBatchGetResponse(raw, "urn:li:sponsoredCreative:586489603")
	if !ok {
		t.Fatal("expected ok")
	}
	if row["name"] != "Ad" {
		t.Fatalf("name = %v", row["name"])
	}
}

func TestParseCreativeBatchGetResponse_missingCreative(t *testing.T) {
	t.Parallel()

	raw := map[string]any{"results": map[string]any{}}
	if _, ok := parseCreativeBatchGetResponse(raw, "urn:li:sponsoredCreative:999"); ok {
		t.Fatal("expected !ok for missing creative")
	}
}

func TestGetCreative_usesBatchGetWithBATCH_GETHeader(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		pages: []any{
			map[string]any{
				"results": map[string]any{
					"urn:li:sponsoredCreative:586489603": map[string]any{
						"id": "urn:li:sponsoredCreative:586489603",
						"content": map[string]any{
							"reference": "urn:li:share:7292289207169933312",
						},
						"leadgenCallToAction": map[string]any{
							"label":       "LEARN_MORE",
							"destination": "urn:li:adForm:12376713",
						},
					},
				},
			},
			map[string]any{
				"results": map[string]any{
					"12376713": map[string]any{"name": "Understory Consultation"},
				},
			},
		},
	}

	svc := &service{proxy: upstream}
	result, err := svc.getCreative(
		context.Background(),
		"user",
		"linkedin_get_creative",
		getCreativeInput{
			accountID:   "512247261",
			creativeURN: "586489603",
			enrichOpts: enrichCreativesOptions{
				includePreviewURLs:     false,
				includeLeadFormDetails: true,
			},
		},
	)
	if err != nil {
		t.Fatalf("getCreative() error = %v", err)
	}

	row, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	if row[fieldFeedURL] != "https://www.linkedin.com/feed/update/urn:li:share:7292289207169933312" {
		t.Fatalf("feedUrl = %v", row[fieldFeedURL])
	}
	if row[fieldLeadFormName] != "Understory Consultation" {
		t.Fatalf("leadFormName = %v", row[fieldLeadFormName])
	}

	if upstream.paths[0] != "adAccounts/512247261/creatives" {
		t.Fatalf("path = %q, want adAccounts/512247261/creatives", upstream.paths[0])
	}
	if upstream.queries[0]["ids"] != "List(urn:li:sponsoredCreative:586489603)" {
		t.Fatalf("ids query = %q", upstream.queries[0]["ids"])
	}
	if upstream.headersList[0]["X-RestLi-Method"] != "BATCH_GET" {
		t.Fatalf("headers = %#v", upstream.headersList[0])
	}
}

func TestGetCreative_notFoundReturnsError(t *testing.T) {
	t.Parallel()

	upstream := &stubLinkedInUpstream{
		pages: []any{map[string]any{"results": map[string]any{}}},
	}
	svc := &service{proxy: upstream}

	_, err := svc.getCreative(
		context.Background(),
		"user",
		"tool",
		getCreativeInput{accountID: "512247261", creativeURN: "999"},
	)
	if err == nil {
		t.Fatal("expected error for missing creative")
	}
}
