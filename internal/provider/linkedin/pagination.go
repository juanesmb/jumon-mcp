package linkedin

import "context"

func fetchSearchPages(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool, apiPath string,
	query map[string]string,
	autoPaginate bool,
	extraHeaders map[string]string,
) (any, error) {
	raw, _, err := fetchSearchPagesWithTruncation(ctx, proxy, userID, mcpTool, apiPath, query, autoPaginate, extraHeaders)
	return raw, err
}

// fetchSearchPagesWithTruncation aggregates search pages and reports whether the pagination cap was hit.
func fetchSearchPagesWithTruncation(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool, apiPath string,
	query map[string]string,
	autoPaginate bool,
	extraHeaders map[string]string,
) (any, bool, error) {
	if !autoPaginate {
		raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", apiPath, query, nil, extraHeaders)
		return raw, false, err
	}

	allElements := make([]any, 0)
	var lastMeta map[string]any
	truncated := false

	for page := range maxAutoPaginatePages {
		raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", apiPath, query, nil, extraHeaders)
		if err != nil {
			return nil, false, err
		}

		pageMap, ok := raw.(map[string]any)
		if !ok {
			return raw, false, nil
		}

		if elements, ok := pageMap["elements"].([]any); ok {
			allElements = append(allElements, elements...)
		}
		if meta, ok := pageMap["metadata"].(map[string]any); ok {
			lastMeta = meta
		}

		nextToken, _ := lastMeta["nextPageToken"].(string)
		if nextToken == "" {
			break
		}
		if page == maxAutoPaginatePages-1 {
			truncated = true
			break
		}
		query["pageToken"] = nextToken
	}

	result := map[string]any{"elements": allElements}
	if lastMeta != nil {
		delete(lastMeta, "nextPageToken")
		result["metadata"] = lastMeta
	}
	return result, truncated, nil
}
