package linkedin

import (
	"context"
	"net/url"
	"strings"
)

// fetchCreativeThumbnailURL resolves a thumbnail URL for a creative's content reference.
// Resolution chain:
//  1. GET /rest/posts/{encodedRef} → extract media URN
//  2. For images:  GET /rest/images/{id}  → downloadUrl
//  3. For videos:  GET /rest/videos/{id}  → downloadUrl (or skip if unavailable)
//
// Returns ("", false) on any error so callers never block.
func fetchCreativeThumbnailURL(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool string,
	contentReference string,
) (string, bool) {
	post, ok := fetchPost(ctx, proxy, userID, mcpTool, contentReference)
	if !ok {
		return "", false
	}
	mediaURN, mediaType := mediaURNFromPost(post)
	if mediaURN == "" {
		return "", false
	}
	switch mediaType {
	case "image":
		return fetchImageDownloadURL(ctx, proxy, userID, mcpTool, mediaURN)
	case "video":
		return fetchVideoThumbnailURL(ctx, proxy, userID, mcpTool, mediaURN)
	default:
		return "", false
	}
}

func fetchPost(ctx context.Context, proxy linkedinUpstreamPort, userID, mcpTool, reference string) (map[string]any, bool) {
	encoded := url.PathEscape(strings.TrimSpace(reference))
	if encoded == "" {
		return nil, false
	}
	raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", "posts/"+encoded, nil, nil, nil)
	if err != nil {
		return nil, false
	}
	post, ok := raw.(map[string]any)
	return post, ok
}

func fetchImageDownloadURL(ctx context.Context, proxy linkedinUpstreamPort, userID, mcpTool, imageURN string) (string, bool) {
	// Strip urn:li:image: prefix for the path segment.
	id := strings.TrimPrefix(imageURN, imageURNPrefix)
	if id == "" {
		return "", false
	}
	raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", "images/"+url.PathEscape(id), nil, nil, nil)
	if err != nil {
		return "", false
	}
	img, ok := raw.(map[string]any)
	if !ok {
		return "", false
	}
	dlURL, _ := img["downloadUrl"].(string)
	dlURL = strings.TrimSpace(dlURL)
	if dlURL == "" {
		return "", false
	}
	return dlURL, true
}

func fetchVideoThumbnailURL(ctx context.Context, proxy linkedinUpstreamPort, userID, mcpTool, videoURN string) (string, bool) {
	id := strings.TrimPrefix(videoURN, videoURNPrefix)
	if id == "" {
		return "", false
	}
	raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", "videos/"+url.PathEscape(id), nil, nil, nil)
	if err != nil {
		return "", false
	}
	vid, ok := raw.(map[string]any)
	if !ok {
		return "", false
	}
	// Prefer explicit downloadUrl if present.
	if dlURL, _ := vid["downloadUrl"].(string); strings.TrimSpace(dlURL) != "" {
		return strings.TrimSpace(dlURL), true
	}
	// Fall back to the video's thumbnail image.
	thumbURN, _ := vid["thumbnail"].(string)
	if strings.TrimSpace(thumbURN) == "" {
		return "", false
	}
	return fetchImageDownloadURL(ctx, proxy, userID, mcpTool, strings.TrimSpace(thumbURN))
}
