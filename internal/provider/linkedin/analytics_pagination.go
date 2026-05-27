package linkedin

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

func applyAnalyticsPagination(query map[string]string, pageToken string, pageSize int) {
	if pageSize > 0 {
		query["count"] = strconv.Itoa(pageSize)
	}
	if start, ok := parseAnalyticsStartToken(pageToken); ok {
		query["start"] = start
	}
}

// parseAnalyticsStartToken accepts a numeric offset, "start=N", or a paging link href.
func parseAnalyticsStartToken(pageToken string) (string, bool) {
	token := strings.TrimSpace(pageToken)
	if token == "" {
		return "", false
	}

	if start, ok := strings.CutPrefix(token, "start="); ok {
		start = strings.TrimSpace(start)
		if start != "" {
			return start, true
		}
	}

	if values, err := url.ParseQuery(token); err == nil {
		if start := strings.TrimSpace(values.Get("start")); start != "" {
			return start, true
		}
	}

	if start := analyticsStartFromHref(token); start != "" {
		return start, true
	}

	if _, err := strconv.Atoi(token); err == nil {
		return token, true
	}

	return "", false
}

func analyticsStartFromHref(href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}

	query := href
	if idx := strings.Index(href, "?"); idx >= 0 {
		query = href[idx+1:]
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(values.Get("start"))
}

func nextAnalyticsStart(raw any) string {
	pageMap, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	paging, ok := pageMap["paging"].(map[string]any)
	if !ok {
		return ""
	}
	links, ok := paging["links"].([]any)
	if !ok {
		return ""
	}
	for _, link := range links {
		linkMap, ok := link.(map[string]any)
		if !ok {
			continue
		}
		if linkMap["rel"] != "next" {
			continue
		}
		href, _ := linkMap["href"].(string)
		return analyticsStartFromHref(href)
	}
	return ""
}

func fetchAnalyticsPages(
	ctx context.Context,
	proxy linkedinUpstreamPort,
	userID, mcpTool string,
	query map[string]string,
	autoPaginate bool,
) (any, error) {
	if !autoPaginate {
		return proxy.requestJSON(ctx, userID, mcpTool, "GET", "adAnalytics", query, nil, nil)
	}

	allElements := make([]any, 0)
	seenPivotKeys := make(map[string]struct{})
	var lastPaging map[string]any

	for range maxAutoPaginatePages {
		raw, err := proxy.requestJSON(ctx, userID, mcpTool, "GET", "adAnalytics", query, nil, nil)
		if err != nil {
			return nil, err
		}

		pageMap, ok := raw.(map[string]any)
		if !ok {
			return raw, nil
		}

		elements, ok := pageMap["elements"].([]any)
		if !ok || len(elements) == 0 {
			break
		}

		added := appendUniqueAnalyticsElements(&allElements, seenPivotKeys, elements)
		if added == 0 {
			break
		}
		if paging, ok := pageMap["paging"].(map[string]any); ok {
			lastPaging = paging
		}

		nextStart := nextAnalyticsStart(raw)
		if nextStart == "" || nextStart == query["start"] {
			break
		}
		query["start"] = nextStart
	}

	result := map[string]any{"elements": allElements}
	if lastPaging != nil {
		result["paging"] = stripNextAnalyticsPaging(lastPaging)
	}
	return result, nil
}

func analyticsPivotKey(element any) string {
	row, ok := element.(map[string]any)
	if !ok {
		return ""
	}
	pivots, ok := row["pivotValues"].([]any)
	if !ok || len(pivots) == 0 {
		return ""
	}
	if pivot, ok := pivots[0].(string); ok {
		return pivot
	}
	return ""
}

func appendUniqueAnalyticsElements(all *[]any, seen map[string]struct{}, page []any) int {
	added := 0
	for _, element := range page {
		key := analyticsPivotKey(element)
		if key == "" {
			*all = append(*all, element)
			added++
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		*all = append(*all, element)
		added++
	}
	return added
}

func stripNextAnalyticsPaging(paging map[string]any) map[string]any {
	links, ok := paging["links"].([]any)
	if !ok {
		return paging
	}

	filtered := make([]any, 0, len(links))
	for _, link := range links {
		linkMap, ok := link.(map[string]any)
		if !ok {
			continue
		}
		if linkMap["rel"] == "next" {
			continue
		}
		filtered = append(filtered, link)
	}

	out := make(map[string]any, len(paging))
	for k, v := range paging {
		out[k] = v
	}
	out["links"] = filtered
	return out
}
