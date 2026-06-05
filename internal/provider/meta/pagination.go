package meta

import (
	"context"
	"fmt"
)

func (s *service) graphGETPaginated(
	ctx context.Context,
	mcpTool, userID, path string,
	baseQuery map[string]string,
	autoPaginate bool,
) (any, error) {
	if !autoPaginate {
		raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, path, baseQuery)
		if err != nil {
			return nil, err
		}
		root, err := unmarshalPayload(raw)
		if err != nil {
			return nil, err
		}
		pageMap, ok := root.(map[string]any)
		if !ok {
			return root, nil
		}
		if _, hasMore := pagingAfterCursor(pageMap); hasMore {
			meta := map[string]any{"has_more": true}
			meta["hint"] = "pass after cursor from paging.cursors.after for the next page"
			pageMap["metadata"] = meta
		}
		return pageMap, nil
	}

	allData := make([]any, 0)
	var merged map[string]any
	query := copyQuery(baseQuery)
	pagesFetched := 0
	truncated := false

	for page := 0; page < maxAutoPaginatePages; page++ {
		raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, path, query)
		if err != nil {
			return nil, err
		}
		pagesFetched++

		root, err := unmarshalPayload(raw)
		if err != nil {
			return nil, err
		}
		pageMap, ok := root.(map[string]any)
		if !ok {
			return root, nil
		}
		if merged == nil {
			merged = make(map[string]any, len(pageMap)+1)
			for k, v := range pageMap {
				if k != "data" && k != "paging" {
					merged[k] = v
				}
			}
		}
		if items, ok := pageMap["data"].([]any); ok {
			allData = append(allData, items...)
		}

		after, hasMore := pagingAfterCursor(pageMap)
		if !hasMore {
			break
		}
		if page == maxAutoPaginatePages-1 {
			truncated = true
			break
		}
		query["after"] = after
	}

	if merged == nil {
		return map[string]any{"data": allData}, nil
	}
	merged["data"] = allData
	if pagesFetched > 1 || truncated {
		meta := map[string]any{
			"pages_fetched": pagesFetched,
			"row_count":     len(allData),
		}
		if truncated {
			meta["truncated"] = true
			meta["hint"] = fmt.Sprintf("auto_paginate stopped after %d pages; narrow filters or set auto_paginate=false and pass after cursor", maxAutoPaginatePages)
		}
		merged["metadata"] = meta
	}
	return merged, nil
}

func pagingAfterCursor(page map[string]any) (string, bool) {
	paging, ok := page["paging"].(map[string]any)
	if !ok {
		return "", false
	}
	cursors, ok := paging["cursors"].(map[string]any)
	if !ok {
		return "", false
	}
	next, _ := paging["next"].(string)
	if next == "" {
		return "", false
	}
	after, _ := cursors["after"].(string)
	return after, after != ""
}

func copyQuery(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
