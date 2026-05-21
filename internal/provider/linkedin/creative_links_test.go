package linkedin

import "testing"

func TestBuildFeedURL_shareAndUgcPost(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"urn:li:share:6778045555198214144":   "https://www.linkedin.com/feed/update/urn:li:share:6778045555198214144",
		"urn:li:ugcPost:6778045555198214144": "https://www.linkedin.com/feed/update/urn:li:ugcPost:6778045555198214144",
	}
	for reference, want := range cases {
		got, ok := buildFeedURL(reference)
		if !ok {
			t.Fatalf("buildFeedURL(%q) = !ok", reference)
		}
		if got != want {
			t.Fatalf("buildFeedURL(%q) = %q, want %q", reference, got, want)
		}
	}
}

func TestBuildFeedURL_rejectsOtherReferences(t *testing.T) {
	t.Parallel()

	if _, ok := buildFeedURL("urn:li:sponsoredCreative:123"); ok {
		t.Fatal("expected sponsoredCreative reference to be rejected")
	}
}

func TestContentReferenceFromCreative(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"content": map[string]any{
			"reference": "urn:li:ugcPost:123",
		},
	}
	got, ok := contentReferenceFromCreative(row)
	if !ok || got != "urn:li:ugcPost:123" {
		t.Fatalf("contentReferenceFromCreative() = %q, %v", got, ok)
	}
}

func TestExtractIframeSrc(t *testing.T) {
	t.Parallel()

	html := `<iframe src='https://www.linkedin.com/ads/ad-preview/?d=abc123' height=580 width=650></iframe>`
	got, ok := extractIframeSrc(html)
	if !ok {
		t.Fatal("extractIframeSrc() = !ok")
	}
	if got != "https://www.linkedin.com/ads/ad-preview/?d=abc123" {
		t.Fatalf("extractIframeSrc() = %q", got)
	}
}

func TestExtractIframeSrc_doubleQuotes(t *testing.T) {
	t.Parallel()

	html := `<iframe src="https://www.linkedin.com/ads/ad-preview/?d=xyz" height=580></iframe>`
	got, ok := extractIframeSrc(html)
	if !ok || got != "https://www.linkedin.com/ads/ad-preview/?d=xyz" {
		t.Fatalf("extractIframeSrc() = %q, %v", got, ok)
	}
}

func TestExtractIframeSrc_invalidHTML(t *testing.T) {
	t.Parallel()

	if _, ok := extractIframeSrc("not an iframe"); ok {
		t.Fatal("expected invalid HTML to miss")
	}
}
