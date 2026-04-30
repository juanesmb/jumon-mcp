package linkedin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	"jumon-mcp/internal/provider/registry"
)

const (
	platformName = "linkedin"
)

func RegisterTools(reg *registry.Registry, gatewayClient *gateway.Client) error {
	tools := []registry.ToolDefinition{
		{
			Name:               "linkedin_list_ad_accounts",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists LinkedIn ad accounts available to the authenticated user.",
			Description:        "Fetches LinkedIn ad accounts with optional filters for status, IDs, names, and pagination.",
			InputSchema:        listAdAccountsSchema(),
			RequiresConnection: true,
			Execute:            listAdAccountsExecutor(gatewayClient),
		},
		{
			Name:               "linkedin_get_campaigns",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches LinkedIn campaigns for one ad account.",
			Description:        "Fetches campaigns with optional status, campaign group, type, name, test, and paging filters.",
			InputSchema:        getCampaignsSchema(),
			RequiresConnection: true,
			Execute:            getCampaignsExecutor(gatewayClient),
		},
		{
			Name:               "linkedin_get_ad_analytics",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches LinkedIn ad analytics by account/campaign grouping.",
			Description:        "Fetches analytics metrics for LinkedIn ads by pivot and date range.",
			InputSchema:        getAnalyticsSchema(),
			RequiresConnection: true,
			Execute:            getAnalyticsExecutor(gatewayClient),
		},
		{
			Name:               "linkedin_search_creatives",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists LinkedIn creatives for selected campaign URNs.",
			Description:        "Fetches creatives via LinkedIn criteria finder for one account and one or more campaign IDs/URNs.",
			InputSchema:        searchCreativesSchema(),
			RequiresConnection: true,
			Execute:            searchCreativesExecutor(gatewayClient),
		},
	}

	for _, tool := range tools {
		if err := reg.Register(tool); err != nil {
			return err
		}
	}
	return nil
}

func listAdAccountsExecutor(gatewayClient *gateway.Client) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		query := map[string]string{"q": "search"}
		appendListFinder(query, "status", toStringSlice(params["status_filter"]))
		appendListFinder(query, "id", toStringSlice(params["account_ids"]))
		appendListFinder(query, "name", toStringSlice(params["name_filter"]))
		appendListFinder(query, "reference", toStringSlice(params["reference_filter"]))
		if testValue, ok := params["test_filter"].(bool); ok {
			query["search.test"] = strconv.FormatBool(testValue)
		}
		if count, ok := toInt(params["page_size"]); ok && count > 0 {
			query["count"] = strconv.Itoa(count)
		}
		if start, ok := toInt(params["start"]); ok && start >= 0 {
			query["start"] = strconv.Itoa(start)
		}

		return proxyLinkedInJSON(ctx, gatewayClient, userID, "GET", "adAccounts", query, nil, nil)
	}
}

func getCampaignsExecutor(gatewayClient *gateway.Client) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		accountID := strings.TrimSpace(toString(params["account_id"]))
		if accountID == "" {
			return nil, fmt.Errorf("account_id is required")
		}
		path := fmt.Sprintf("adAccounts/%s/adCampaigns", url.PathEscape(accountID))
		query := map[string]string{"q": "search"}
		appendListFinder(query, "status", toStringSlice(params["status_filter"]))
		appendListFinder(query, "campaignGroup", toCampaignGroupURNs(params["campaign_group_filter"]))
		appendListFinder(query, "type", toStringSlice(params["type_filter"]))
		appendListFinder(query, "name", toStringSlice(params["name_filter"]))
		if testValue, ok := params["test_filter"].(bool); ok {
			query["search"] = mergeSearchItem(query["search"], fmt.Sprintf("test:%t", testValue))
		}
		if sortOrder := strings.TrimSpace(toString(params["sort_order"])); sortOrder != "" {
			query["sortOrder"] = sortOrder
		}
		if pageSize, ok := toInt(params["page_size"]); ok && pageSize > 0 {
			query["pageSize"] = strconv.Itoa(pageSize)
		}
		if token := strings.TrimSpace(toString(params["page_token"])); token != "" {
			query["pageToken"] = token
		}
		return proxyLinkedInJSON(ctx, gatewayClient, userID, "GET", path, query, nil, nil)
	}
}

func getAnalyticsExecutor(gatewayClient *gateway.Client) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		accountID := strings.TrimSpace(toString(params["account_id"]))
		if accountID == "" {
			return nil, fmt.Errorf("account_id is required")
		}
		startDate := strings.TrimSpace(toString(params["date_range_start"]))
		if startDate == "" {
			return nil, fmt.Errorf("date_range_start is required")
		}
		start, err := parseDate(startDate)
		if err != nil {
			return nil, err
		}

		finderType := strings.TrimSpace(toString(params["finder_type"]))
		if finderType == "" {
			finderType = "analytics"
		}
		query := map[string]string{
			"q":        finderType,
			"accounts": fmt.Sprintf("List(%s)", url.QueryEscape("urn:li:sponsoredAccount:"+accountID)),
			"dateRange": fmt.Sprintf(
				"(start:(year:%d,month:%d,day:%d)%s)",
				start.Year(), int(start.Month()), start.Day(), buildDateRangeEnd(params["date_range_end"]),
			),
		}

		pivots := toStringSlice(params["pivots"])
		if len(pivots) > 0 {
			query["pivot"] = strings.TrimSpace(pivots[0])
		}
		if granularity := strings.TrimSpace(toString(params["time_granularity"])); granularity != "" {
			query["timeGranularity"] = granularity
		}
		if fields := toStringSlice(params["fields"]); len(fields) > 0 {
			query["fields"] = strings.Join(fields, ",")
		}
		if campaignIDs := toStringSlice(params["campaign_ids"]); len(campaignIDs) > 0 {
			query["campaigns"] = fmt.Sprintf("List(%s)", strings.Join(toURNList("urn:li:sponsoredCampaign:", campaignIDs), ","))
		}
		if creativeIDs := toStringSlice(params["creative_ids"]); len(creativeIDs) > 0 {
			query["creatives"] = fmt.Sprintf("List(%s)", strings.Join(toURNList("urn:li:sponsoredCreative:", creativeIDs), ","))
		}
		if groups := toStringSlice(params["campaign_group_ids"]); len(groups) > 0 {
			query["campaignGroups"] = fmt.Sprintf("List(%s)", strings.Join(toURNList("urn:li:sponsoredCampaignGroup:", groups), ","))
		}
		if sortField := strings.TrimSpace(toString(params["sort_by_field"])); sortField != "" {
			sortOrder := strings.TrimSpace(toString(params["sort_by_order"]))
			if sortOrder == "" {
				sortOrder = "DESCENDING"
			}
			query["sortBy"] = fmt.Sprintf("(field:%s,order:%s)", sortField, sortOrder)
		}

		return proxyLinkedInJSON(ctx, gatewayClient, userID, "GET", "adAnalytics", query, nil, nil)
	}
}

func searchCreativesExecutor(gatewayClient *gateway.Client) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		accountID := strings.TrimSpace(toString(params["account_id"]))
		if accountID == "" {
			return nil, fmt.Errorf("account_id is required")
		}

		campaignIDs := toStringSlice(params["campaign_ids"])
		if len(campaignIDs) == 0 {
			campaignIDs = toStringSlice(params["campaign_urns"])
		}
		if len(campaignIDs) == 0 {
			return nil, fmt.Errorf("campaign_ids or campaign_urns is required")
		}

		query := map[string]string{
			"q":         "criteria",
			"campaigns": fmt.Sprintf("List(%s)", strings.Join(toCampaignURNs(campaignIDs), ",")),
		}
		if sortOrder := strings.TrimSpace(toString(params["sort_order"])); sortOrder != "" {
			query["sortOrder"] = sortOrder
		}
		if pageSize, ok := toInt(params["page_size"]); ok && pageSize > 0 {
			query["pageSize"] = strconv.Itoa(pageSize)
		}
		if pageToken := strings.TrimSpace(toString(params["page_token"])); pageToken != "" {
			query["pageToken"] = pageToken
		}

		headers := map[string]string{"X-RestLi-Method": "FINDER"}
		path := fmt.Sprintf("adAccounts/%s/creatives", url.PathEscape(accountID))
		return proxyLinkedInJSON(ctx, gatewayClient, userID, "GET", path, query, nil, headers)
	}
}

func proxyLinkedInJSON(ctx context.Context, gatewayClient *gateway.Client, userID, method, path string, query map[string]string, body any, extraHeaders map[string]string) (any, error) {
	headers := map[string]string{
		"Linkedin-Version":         "202504",
		"X-Restli-Protocol-Version": "2.0.0",
	}
	for key, value := range extraHeaders {
		headers[key] = value
	}
	resp, err := gatewayClient.ProxyProviderOrRefresh(ctx, platformName, userID, method, path, query, body, headers)
	if err != nil {
		return nil, err
	}

	if gateway.IsNotConnectedResponse(resp) {
		return nil, &catalog.PlatformNotConnectedError{Platform: platformName, ConnectURL: gatewayClient.ConnectURLHint()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("linkedin api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(resp.Body)))
	}
	var payload any
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return map[string]any{
			"status": resp.StatusCode,
			"body":   string(resp.Body),
		}, nil
	}
	return payload, nil
}

func appendListFinder(query map[string]string, field string, values []string) {
	if len(values) == 0 {
		return
	}
	encoded := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		encoded = append(encoded, trimmed)
	}
	if len(encoded) == 0 {
		return
	}
	searchItem := fmt.Sprintf("%s:(values:List(%s))", field, strings.Join(encoded, ","))
	query["search"] = mergeSearchItem(query["search"], searchItem)
}

func mergeSearchItem(existing, next string) string {
	trimmedExisting := strings.TrimSpace(existing)
	trimmedNext := strings.TrimSpace(next)
	if trimmedExisting == "" {
		return fmt.Sprintf("(%s)", trimmedNext)
	}
	raw := strings.TrimPrefix(trimmedExisting, "(")
	raw = strings.TrimSuffix(raw, ")")
	if raw == "" {
		return fmt.Sprintf("(%s)", trimmedNext)
	}
	return fmt.Sprintf("(%s,%s)", raw, trimmedNext)
}

func toCampaignGroupURNs(value any) []string {
	items := toStringSlice(value)
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "urn:li:sponsoredCampaignGroup:") {
			out = append(out, trimmed)
			continue
		}
		out = append(out, "urn:li:sponsoredCampaignGroup:"+trimmed)
	}
	return out
}

func toCampaignURNs(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "urn:li:sponsoredCampaign:") {
			out = append(out, url.QueryEscape(trimmed))
			continue
		}
		out = append(out, url.QueryEscape("urn:li:sponsoredCampaign:"+trimmed))
	}
	return out
}

func toURNList(prefix string, items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "urn:") {
			out = append(out, url.QueryEscape(trimmed))
		} else {
			out = append(out, url.QueryEscape(prefix+trimmed))
		}
	}
	return out
}

func buildDateRangeEnd(raw any) string {
	value := strings.TrimSpace(toString(raw))
	if value == "" {
		return ""
	}
	endDate, err := parseDate(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf(",end:(year:%d,month:%d,day:%d)", endDate.Year(), int(endDate.Month()), endDate.Day())
}

func parseDate(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", value)
	}
	return date, nil
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func toStringSlice(value any) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{v}
	default:
		return nil
	}
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

func listAdAccountsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status_filter":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"account_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"name_filter":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"reference_filter": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"test_filter":      map[string]any{"type": "boolean"},
			"page_size":        map[string]any{"type": "number"},
			"start":            map[string]any{"type": "number"},
		},
	}
}

func getCampaignsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{
			"account_id",
		},
		"properties": map[string]any{
			"account_id":            map[string]any{"type": "string"},
			"campaign_group_filter": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"status_filter":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"type_filter":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"name_filter":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"test_filter":           map[string]any{"type": "boolean"},
			"sort_order":            map[string]any{"type": "string", "enum": []string{"ASCENDING", "DESCENDING"}},
			"page_size":             map[string]any{"type": "number"},
			"page_token":            map[string]any{"type": "string"},
		},
	}
}

func getAnalyticsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{
			"account_id",
			"date_range_start",
			"pivots",
		},
		"properties": map[string]any{
			"finder_type":        map[string]any{"type": "string", "enum": []string{"analytics", "statistics", "attributedRevenueMetrics"}},
			"pivots":             map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"date_range_start":   map[string]any{"type": "string"},
			"date_range_end":     map[string]any{"type": "string"},
			"time_granularity":   map[string]any{"type": "string", "enum": []string{"ALL", "DAILY", "MONTHLY", "YEARLY"}},
			"account_id":         map[string]any{"type": "string"},
			"campaign_group_ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"campaign_ids":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"creative_ids":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"fields":             map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"sort_by_field":      map[string]any{"type": "string"},
			"sort_by_order":      map[string]any{"type": "string", "enum": []string{"ASCENDING", "DESCENDING"}},
		},
	}
}

func searchCreativesSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{
			"account_id",
		},
		"properties": map[string]any{
			"account_id":    map[string]any{"type": "string"},
			"campaign_ids":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"campaign_urns": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"sort_order":    map[string]any{"type": "string", "enum": []string{"ASCENDING", "DESCENDING"}},
			"page_size":     map[string]any{"type": "number"},
			"page_token":    map[string]any{"type": "string"},
		},
	}
}
