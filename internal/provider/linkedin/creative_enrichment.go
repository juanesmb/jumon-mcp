package linkedin

import "context"

type enrichCreativesOptions struct {
	includePreviewURLs     bool
	includeLeadFormDetails bool
	includeAssetURLs       bool
}

func enrichCreativesResponse(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool, accountID string,
	payload any,
	opts enrichCreativesOptions,
) any {
	root, ok := payload.(map[string]any)
	if !ok {
		return payload
	}

	elements, ok := root["elements"].([]any)
	if !ok {
		return payload
	}

	for i, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		enrichCreativeElement(ctx, proxy, userID, mcpTool, accountID, row, opts)
		elements[i] = row
	}

	if opts.includeLeadFormDetails {
		enrichCreativesWithLeadForms(ctx, proxy, userID, mcpTool, elements)
	}

	if opts.includeAssetURLs {
		enrichCreativesWithAssets(ctx, proxy, userID, mcpTool, elements)
	}

	root["elements"] = elements
	return root
}

func enrichCreativeElement(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool, accountID string,
	row map[string]any,
	opts enrichCreativesOptions,
) {
	if reference, ok := contentReferenceFromCreative(row); ok {
		if feedURL, ok := buildFeedURL(reference); ok {
			row[fieldFeedURL] = feedURL
		}
	}

	if !opts.includePreviewURLs {
		return
	}

	creativeURN, ok := creativeURNFromRow(row)
	if !ok {
		return
	}
	if previewURL, ok := fetchCreativePreviewURL(ctx, proxy, userID, mcpTool, accountID, creativeURN); ok {
		row[fieldPreviewURL] = previewURL
	}
}
