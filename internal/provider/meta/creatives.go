package meta

import (
	"context"
	"fmt"
	"strings"
)

func (s *service) listCreatives(ctx context.Context, mcpTool, userID, actID string, in listCreativesInput) (any, error) {
	normalized, err := normalizeActID(actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in.listPaginationInput)
	if len(in.filtering) > 0 {
		query["filtering"] = jsonEncode(in.filtering)
	}
	return s.graphGETPaginated(ctx, mcpTool, userID, normalized+"/adcreatives", query, in.autoPaginate)
}

func (s *service) getCreative(ctx context.Context, mcpTool, userID string, in getCreativeInput) (any, error) {
	creativeID := strings.TrimSpace(in.creativeID)
	if creativeID == "" {
		return nil, fmt.Errorf("meta: creative_id is required")
	}
	fields := in.fields
	if len(fields) == 0 {
		fields = append([]string(nil), defaultCreativeFields...)
	}
	query := map[string]string{"fields": joinCSV(fields)}
	if in.thumbnailWidth > 0 {
		query["thumbnail_width"] = toString(in.thumbnailWidth)
	}
	if in.thumbnailHeight > 0 {
		query["thumbnail_height"] = toString(in.thumbnailHeight)
	}
	raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, creativeID, query)
	if err != nil {
		return nil, err
	}
	return unmarshalPayload(raw)
}

type listCreativesInput struct {
	listPaginationInput
	filtering []map[string]any
}

type getCreativeInput struct {
	creativeID      string
	fields          []string
	thumbnailWidth  int
	thumbnailHeight int
}

func parseListCreativesInput(params map[string]any) (listCreativesInput, error) {
	filtering, err := parseFiltering(params["filtering"])
	if err != nil {
		return listCreativesInput{}, err
	}
	return listCreativesInput{
		listPaginationInput: parseListPagination(params, defaultCreativeListFields),
		filtering:           filtering,
	}, nil
}

func (s *service) listCreativeAds(ctx context.Context, mcpTool, userID string, in listCreativeAdsInput) (any, error) {
	normalized, err := normalizeActID(in.actID)
	if err != nil {
		return nil, err
	}
	query := buildListQuery(in.listPaginationInput)
	path := normalized + "/ads"

	if !in.autoPaginate {
		raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, path, query)
		if err != nil {
			return nil, err
		}
		page, err := unmarshalPayload(raw)
		if err != nil {
			return nil, err
		}
		return filterCreativeAdsPage(page, in.creativeID), nil
	}

	matches := make([]any, 0)
	pagesFetched := 0
	truncated := false
	pageQuery := copyQuery(query)

	for page := 0; page < maxAutoPaginatePages; page++ {
		raw, err := s.proxy.getWithRefresh(ctx, mcpTool, userID, path, pageQuery)
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
		if items, ok := pageMap["data"].([]any); ok {
			matches = append(matches, filterAdsByCreativeID(items, in.creativeID)...)
		}

		after, hasMore := pagingAfterCursor(pageMap)
		if !hasMore {
			break
		}
		if page == maxAutoPaginatePages-1 {
			truncated = true
			break
		}
		pageQuery["after"] = after
	}

	out := map[string]any{"data": matches}
	if pagesFetched > 1 || truncated {
		meta := map[string]any{
			"pages_fetched": pagesFetched,
			"row_count":     len(matches),
		}
		if truncated {
			meta["truncated"] = true
			meta["hint"] = fmt.Sprintf("auto_paginate stopped after %d pages; set auto_paginate=false and pass after cursor on meta_list_ads if more matches may exist", maxAutoPaginatePages)
		}
		out["metadata"] = meta
	}
	return out, nil
}

func filterCreativeAdsPage(page any, creativeID string) any {
	pageMap, ok := page.(map[string]any)
	if !ok {
		return page
	}
	if items, ok := pageMap["data"].([]any); ok {
		pageMap["data"] = filterAdsByCreativeID(items, creativeID)
	}
	if _, hasMore := pagingAfterCursor(pageMap); hasMore {
		meta := map[string]any{"has_more": true}
		meta["hint"] = "only the current page was scanned; matching ads on later pages are omitted unless auto_paginate=true"
		pageMap["metadata"] = meta
	}
	return pageMap
}

func filterAdsByCreativeID(ads []any, creativeID string) []any {
	filtered := make([]any, 0, len(ads))
	for _, item := range ads {
		ad, ok := item.(map[string]any)
		if !ok || !adUsesCreative(ad, creativeID) {
			continue
		}
		filtered = append(filtered, ad)
	}
	return filtered
}

func adUsesCreative(ad map[string]any, creativeID string) bool {
	creative, ok := ad["creative"].(map[string]any)
	if !ok {
		return false
	}
	return toString(creative["id"]) == creativeID
}

func parseGetCreativeInput(params map[string]any) getCreativeInput {
	in := getCreativeInput{
		creativeID: strings.TrimSpace(toString(params["creative_id"])),
		fields:     toStringSlice(params["fields"]),
	}
	if w, ok := toInt(params["thumbnail_width"]); ok {
		in.thumbnailWidth = w
	}
	if h, ok := toInt(params["thumbnail_height"]); ok {
		in.thumbnailHeight = h
	}
	return in
}
