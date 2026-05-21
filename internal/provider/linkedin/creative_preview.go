package linkedin

import (
	"context"
	"strings"
)

func fetchCreativePreviewURL(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool, accountID, creativeURN string,
) (string, bool) {
	accountURN := sponsoredAccountURN(accountID)
	query := map[string]string{
		"q":        "creative",
		"creative": creativeURN,
		"account":  accountURN,
	}

	raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", "adPreviews", query, nil, nil)
	if err != nil {
		return "", false
	}

	previewHTML, ok := selectDesktopFeedPreviewHTML(raw)
	if !ok {
		return "", false
	}
	src, ok := extractIframeSrc(previewHTML)
	if !ok {
		return "", false
	}
	return src, true
}

func previewHTMLFromElement(item any) (string, bool) {
	row, ok := item.(map[string]any)
	if !ok {
		return "", false
	}
	preview, ok := row["preview"].(string)
	if !ok || strings.TrimSpace(preview) == "" {
		return "", false
	}
	return preview, true
}

func selectDesktopFeedPreviewHTML(raw any) (string, bool) {
	root, ok := raw.(map[string]any)
	if !ok {
		return "", false
	}
	elements, ok := root["elements"].([]any)
	if !ok || len(elements) == 0 {
		return "", false
	}

	for _, item := range elements {
		row, ok := item.(map[string]any)
		if !ok || !isDesktopFeedPlacement(row) {
			continue
		}
		if preview, ok := previewHTMLFromElement(row); ok {
			return preview, true
		}
	}

	for _, item := range elements {
		if preview, ok := previewHTMLFromElement(item); ok {
			return preview, true
		}
	}
	return "", false
}

func isDesktopFeedPlacement(row map[string]any) bool {
	placement, ok := row["placement"].(map[string]any)
	if !ok {
		return false
	}
	linkedinPlacement, ok := placement["linkedin"].(map[string]any)
	if !ok {
		return false
	}
	name, _ := linkedinPlacement["placementName"].(string)
	presentation, _ := linkedinPlacement["contentPresentationType"].(string)
	return strings.EqualFold(name, "FEED") && strings.EqualFold(presentation, "DESKTOP_WEBSITE")
}
