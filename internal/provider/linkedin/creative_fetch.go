package linkedin

import "strings"

// parseCreativeBatchGetResponse extracts one creative from a BATCH_GET response
// (results map keyed by sponsoredCreative URN). Returns false when the creative
// is absent or the payload shape is unexpected.
func parseCreativeBatchGetResponse(raw any, creativeURN string) (map[string]any, bool) {
	root, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}

	results, ok := root["results"].(map[string]any)
	if !ok {
		return nil, false
	}

	// LinkedIn keys results by the full URN string.
	if item, ok := results[creativeURN]; ok {
		if row, ok := item.(map[string]any); ok {
			return row, true
		}
	}

	// Fall back to matching any key that ends with the numeric creative ID.
	suffix := strings.TrimPrefix(creativeURN, "urn:li:sponsoredCreative:")
	if suffix == creativeURN {
		return nil, false
	}
	for key, item := range results {
		if strings.HasSuffix(key, suffix) {
			if row, ok := item.(map[string]any); ok {
				return row, true
			}
		}
	}
	return nil, false
}
