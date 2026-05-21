package linkedin

import (
	"regexp"
	"strings"
)

const (
	fieldFeedURL    = "feedUrl"
	fieldPreviewURL = "previewUrl"

	linkedInFeedUpdateBase = "https://www.linkedin.com/feed/update/"
)

var iframeSrcPattern = regexp.MustCompile(`(?i)<iframe[^>]+src=['"]([^'"]+)['"]`)

func contentReferenceFromCreative(row map[string]any) (string, bool) {
	content, ok := row["content"].(map[string]any)
	if !ok {
		return "", false
	}
	reference, ok := content["reference"].(string)
	if !ok {
		return "", false
	}
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return "", false
	}
	return reference, true
}

func buildFeedURL(reference string) (string, bool) {
	trimmed := strings.TrimSpace(reference)
	if !strings.HasPrefix(trimmed, "urn:li:share:") && !strings.HasPrefix(trimmed, "urn:li:ugcPost:") {
		return "", false
	}
	return linkedInFeedUpdateBase + trimmed, true
}

func extractIframeSrc(previewHTML string) (string, bool) {
	match := iframeSrcPattern.FindStringSubmatch(previewHTML)
	if len(match) < 2 {
		return "", false
	}
	src := strings.TrimSpace(match[1])
	if src == "" {
		return "", false
	}
	return src, true
}

func creativeURNFromRow(row map[string]any) (string, bool) {
	id, ok := row["id"].(string)
	if !ok {
		return "", false
	}
	id = strings.TrimSpace(id)
	if id == "" || !strings.HasPrefix(id, "urn:li:sponsoredCreative:") {
		return "", false
	}
	return id, true
}

func sponsoredAccountURN(accountID string) string {
	trimmed := strings.TrimSpace(accountID)
	if strings.HasPrefix(trimmed, "urn:li:sponsoredAccount:") {
		return trimmed
	}
	return "urn:li:sponsoredAccount:" + trimmed
}
