package linkedin

import "context"

// enrichCreativesWithAssets adds a thumbnailUrl field to creative elements by resolving
// the content.reference post → image/video chain. Only creatives whose content.reference
// is a share or ugcPost URN are processed. All failures are silently ignored.
//
// Requires the `r_ads` scope to read sponsored posts/images; the enrichment degrades
// gracefully (thumbnailUrl omitted) if the posts or images endpoint returns an error.
func enrichCreativesWithAssets(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool string,
	elements []any,
) {
	referenceForIndex := make(map[int]string, len(elements))
	for i, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ref, ok := contentReferenceFromCreative(row)
		if !ok {
			continue
		}
		referenceForIndex[i] = ref
	}

	if len(referenceForIndex) == 0 {
		return
	}

	refs := make([]string, 0, len(referenceForIndex))
	for _, ref := range referenceForIndex {
		refs = append(refs, ref)
	}
	uniqueRefs := uniqueStrings(refs)

	// Resolve thumbnail URL for each unique reference.
	thumbByRef := make(map[string]string, len(uniqueRefs))
	for _, ref := range uniqueRefs {
		if thumbURL, ok := fetchCreativeThumbnailURL(ctx, proxy, userID, mcpTool, ref); ok {
			thumbByRef[ref] = thumbURL
		}
	}

	// Apply resolved URLs.
	for i, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ref, ok := referenceForIndex[i]
		if !ok {
			continue
		}
		if thumbURL, ok := thumbByRef[ref]; ok {
			row[fieldThumbnailURL] = thumbURL
		}
	}
}
