package linkedin

import "strings"

const (
	fieldThumbnailURL = "thumbnailUrl"

	imageURNPrefix = "urn:li:image:"
	videoURNPrefix = "urn:li:video:"
)

// mediaURNFromPost extracts the image or video URN from a LinkedIn post JSON object.
// It inspects both the newer Posts API shape (content.media.id) and the older UGC Posts
// shape (specificContent.com.linkedin.ugc.ShareContent.media[0].media).
// Returns ("", "") when no recognisable media is found.
func mediaURNFromPost(post map[string]any) (mediaURN, mediaType string) {
	// Newer Posts API: content.media.id
	if content, ok := post["content"].(map[string]any); ok {
		if media, ok := content["media"].(map[string]any); ok {
			if id, ok := media["id"].(string); ok && strings.TrimSpace(id) != "" {
				id = strings.TrimSpace(id)
				return id, mediaTypeFromURN(id)
			}
		}
	}

	// Older UGC Posts API: specificContent.com.linkedin.ugc.ShareContent.media[0].media
	if specific, ok := post["specificContent"].(map[string]any); ok {
		for _, v := range specific {
			content, ok := v.(map[string]any)
			if !ok {
				continue
			}
			mediaSlice, ok := content["media"].([]any)
			if !ok || len(mediaSlice) == 0 {
				continue
			}
			first, ok := mediaSlice[0].(map[string]any)
			if !ok {
				continue
			}
			if urn, ok := first["media"].(string); ok && strings.TrimSpace(urn) != "" {
				urn = strings.TrimSpace(urn)
				return urn, mediaTypeFromURN(urn)
			}
		}
	}

	return "", ""
}

func mediaTypeFromURN(urn string) string {
	switch {
	case strings.HasPrefix(urn, imageURNPrefix):
		return "image"
	case strings.HasPrefix(urn, videoURNPrefix):
		return "video"
	default:
		return "unknown"
	}
}
