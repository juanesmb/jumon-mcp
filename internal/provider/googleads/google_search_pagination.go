package googleads

import (
	"context"
)

const maxAutoPaginatePages = 10

func (s *service) googleSearch(
	ctx context.Context,
	userID, mcpTool,
	customerID, loginCustomerID,
	query string,
) (any, error) {
	return s.googleSearchSinglePage(ctx, userID, mcpTool, customerID, loginCustomerID, query, "")
}

func (s *service) googleSearchPaginated(
	ctx context.Context,
	userID, mcpTool,
	customerID, loginCustomerID,
	query string,
	autoPaginate bool,
) (any, error) {
	if !autoPaginate {
		return s.googleSearchSinglePage(ctx, userID, mcpTool, customerID, loginCustomerID, query, "")
	}

	allResults := make([]any, 0)
	var merged map[string]any
	pageToken := ""
	pagesFetched := 0

	for page := 0; page < maxAutoPaginatePages; page++ {
		raw, err := s.googleSearchSinglePage(ctx, userID, mcpTool, customerID, loginCustomerID, query, pageToken)
		if err != nil {
			return nil, err
		}
		pagesFetched++
		root, ok := raw.(map[string]any)
		if !ok {
			return raw, nil
		}
		if merged == nil {
			merged = make(map[string]any, len(root)+1)
			for k, v := range root {
				if k != "results" && k != "nextPageToken" {
					merged[k] = v
				}
			}
		}
		if items, ok := root["results"].([]any); ok {
			allResults = append(allResults, items...)
		}
		nextToken, _ := root["nextPageToken"].(string)
		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	if merged == nil {
		return map[string]any{"results": allResults}, nil
	}
	merged["results"] = allResults
	if pagesFetched > 1 {
		merged["metadata"] = map[string]any{
			"pages_fetched": pagesFetched,
			"row_count":     len(allResults),
		}
	}
	return merged, nil
}

func (s *service) googleSearchSinglePage(
	ctx context.Context,
	userID, mcpTool,
	customerID, loginCustomerID,
	query, pageToken string,
) (any, error) {
	path := pathGoogleAdsSearch(s.apiVersion, customerID)
	headers := map[string]string{}
	if loginCustomerID != "" {
		headers["login-customer-id"] = loginCustomerID
	}
	body := map[string]any{
		"query": query,
		"searchSettings": map[string]any{
			"returnTotalResultsCount": true,
		},
	}
	if pageToken != "" {
		body["pageToken"] = pageToken
	}
	return s.proxy.requestJSON(ctx, userID, mcpTool, "POST", path, body, headers)
}
