package linkedin

import "context"

func fetchSearchPages(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool, apiPath string,
	query map[string]string,
	autoPaginate bool,
) (any, error) {
	if !autoPaginate {
		return proxy.requestJSON(ctx, userID, mcpTool, "GET", apiPath, query, nil, nil)
	}

	allElements := make([]any, 0)
	var lastMeta map[string]any

	for range maxAutoPaginatePages {
		raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", apiPath, query, nil, nil)
		if err != nil {
			return nil, err
		}

		pageMap, ok := raw.(map[string]any)
		if !ok {
			return raw, nil
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
		query["pageToken"] = nextToken
	}

	result := map[string]any{"elements": allElements}
	if lastMeta != nil {
		delete(lastMeta, "nextPageToken")
		result["metadata"] = lastMeta
	}
	return result, nil
}
