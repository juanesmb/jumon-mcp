package linkedin

import "testing"

func TestSelectDesktopFeedPreviewHTML_prefersDesktopFeed(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"elements": []any{
			map[string]any{
				"preview": "<iframe src='https://www.linkedin.com/mwlite/ads/ad-preview?d=mobile'></iframe>",
				"placement": map[string]any{
					"linkedin": map[string]any{
						"placementName":           "FEED",
						"contentPresentationType": "MOBILE_WEBSITE",
					},
				},
			},
			map[string]any{
				"preview": "<iframe src='https://www.linkedin.com/ads/ad-preview/?d=desktop'></iframe>",
				"placement": map[string]any{
					"linkedin": map[string]any{
						"placementName":           "FEED",
						"contentPresentationType": "DESKTOP_WEBSITE",
					},
				},
			},
		},
	}

	got, ok := selectDesktopFeedPreviewHTML(raw)
	if !ok {
		t.Fatal("selectDesktopFeedPreviewHTML() = !ok")
	}
	src, ok := extractIframeSrc(got)
	if !ok || src != "https://www.linkedin.com/ads/ad-preview/?d=desktop" {
		t.Fatalf("preview = %q, %v", src, ok)
	}
}

func TestSelectDesktopFeedPreviewHTML_fallsBackToFirst(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"elements": []any{
			map[string]any{
				"preview": "<iframe src='https://www.linkedin.com/ads/ad-preview/?d=fallback'></iframe>",
			},
		},
	}

	got, ok := selectDesktopFeedPreviewHTML(raw)
	if !ok {
		t.Fatal("selectDesktopFeedPreviewHTML() = !ok")
	}
	src, ok := extractIframeSrc(got)
	if !ok || src != "https://www.linkedin.com/ads/ad-preview/?d=fallback" {
		t.Fatalf("preview = %q, %v", src, ok)
	}
}

func TestSelectDesktopFeedPreviewHTML_empty(t *testing.T) {
	t.Parallel()

	if _, ok := selectDesktopFeedPreviewHTML(map[string]any{}); ok {
		t.Fatal("expected empty payload to miss")
	}
}
