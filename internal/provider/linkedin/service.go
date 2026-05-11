package linkedin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	defaultCampaignsPageSize = 100
	maxAutoPaginatePages     = 20
)

type service struct {
	proxy linkedinUpstreamPort
}

func newLinkedInService(proxy linkedinUpstreamPort) *service {
	return &service{proxy: proxy}
}

type listAdAccountsInput struct {
	statusFilter    []string
	accountIDs      []string
	nameFilter      []string
	referenceFilter []string
	testFilter      *bool
	pageSize        *int
	start           *int
}

type getCampaignsInput struct {
	accountID      string
	statusFilter   []string
	campaignGroups []string
	typeFilter     []string
	nameFilter     []string
	testFilter     *bool
	sortOrder      string
	autoPaginate   bool
	pageSize       int
	pageToken      string
}

type getAnalyticsInput struct {
	accountID        string
	startDate        string
	endDate          string
	finderType       string
	pivots           []string
	timeGranularity  string
	campaignIDs      []string
	creativeIDs      []string
	campaignGroupIDs []string
	fields           []string
	sortByField      string
	sortByOrder      string
}

type searchCreativesInput struct {
	accountID   string
	campaignIDs []string
	sortOrder   string
	pageSize    *int
	pageToken   string
}

func (s *service) listAdAccounts(ctx context.Context, userID, mcpTool string, in listAdAccountsInput) (any, error) {
	query := map[string]string{"q": "search"}
	appendListFinder(query, "status", in.statusFilter)
	appendListFinder(query, "id", in.accountIDs)
	appendListFinder(query, "name", in.nameFilter)
	appendListFinder(query, "reference", in.referenceFilter)
	if in.testFilter != nil {
		query["search.test"] = strconv.FormatBool(*in.testFilter)
	}
	if in.pageSize != nil && *in.pageSize > 0 {
		query["count"] = strconv.Itoa(*in.pageSize)
	}
	if in.start != nil && *in.start >= 0 {
		query["start"] = strconv.Itoa(*in.start)
	}

	return s.proxy.requestJSON(ctx, userID, mcpTool, "GET", "adAccounts", query, nil, nil)
}

func (s *service) getCampaigns(ctx context.Context, userID, mcpTool string, in getCampaignsInput) (any, error) {
	apiPath := fmt.Sprintf("adAccounts/%s/adCampaigns", url.PathEscape(in.accountID))
	query := map[string]string{"q": "search"}
	appendListFinder(query, "status", in.statusFilter)
	appendListFinder(query, "campaignGroup", toCampaignGroupURNs(in.campaignGroups))
	appendListFinder(query, "type", in.typeFilter)
	appendListFinder(query, "name", in.nameFilter)
	if in.testFilter != nil {
		query["search"] = mergeSearchItem(query["search"], fmt.Sprintf("test:%t", *in.testFilter))
	}
	if strings.TrimSpace(in.sortOrder) != "" {
		query["sortOrder"] = strings.TrimSpace(in.sortOrder)
	}

	query["pageSize"] = strconv.Itoa(in.pageSize)

	// When the caller provides an explicit page_token they want one page only.
	if strings.TrimSpace(in.pageToken) != "" {
		query["pageToken"] = strings.TrimSpace(in.pageToken)
		in.autoPaginate = false
	}

	if !in.autoPaginate {
		return s.proxy.requestJSON(ctx, userID, mcpTool, "GET", apiPath, query, nil, nil)
	}

	allElements := make([]any, 0)
	var lastMeta map[string]any

	for range maxAutoPaginatePages {
		raw, err := s.proxy.requestJSON(ctx, userID, mcpTool, "GET", apiPath, query, nil, nil)
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

func (s *service) getAnalytics(ctx context.Context, userID, mcpTool string, in getAnalyticsInput) (any, error) {
	accountID := strings.TrimSpace(in.accountID)
	startDate := strings.TrimSpace(in.startDate)
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required")
	}
	if startDate == "" {
		return nil, fmt.Errorf("date_range_start is required")
	}

	start, err := parseDate(startDate)
	if err != nil {
		return nil, err
	}

	finderType := strings.TrimSpace(in.finderType)
	if finderType == "" {
		finderType = "analytics"
	}

	query := map[string]string{
		"q":        finderType,
		"accounts": fmt.Sprintf("List(%s)", url.QueryEscape("urn:li:sponsoredAccount:"+accountID)),
		"dateRange": fmt.Sprintf(
			"(start:(year:%d,month:%d,day:%d)%s)",
			start.Year(), int(start.Month()), start.Day(), buildDateRangeEnd(in.endDate),
		),
	}

	if pivots := in.pivots; len(pivots) > 0 {
		query["pivot"] = strings.TrimSpace(pivots[0])
	}
	if granularity := strings.TrimSpace(in.timeGranularity); granularity != "" {
		query["timeGranularity"] = granularity
	}
	if fields := in.fields; len(fields) > 0 {
		if len(in.pivots) > 0 && !slices.Contains(fields, "pivotValues") {
			fields = append([]string{"pivotValues"}, fields...)
		}
		query["fields"] = strings.Join(fields, ",")
	}

	if campaignIDs := in.campaignIDs; len(campaignIDs) > 0 {
		query["campaigns"] = fmt.Sprintf("List(%s)", strings.Join(toURNList("urn:li:sponsoredCampaign:", campaignIDs), ","))
	}
	if creativeIDs := in.creativeIDs; len(creativeIDs) > 0 {
		query["creatives"] = fmt.Sprintf("List(%s)", strings.Join(toURNList("urn:li:sponsoredCreative:", creativeIDs), ","))
	}
	if groups := in.campaignGroupIDs; len(groups) > 0 {
		query["campaignGroups"] = fmt.Sprintf("List(%s)", strings.Join(toURNList("urn:li:sponsoredCampaignGroup:", groups), ","))
	}

	if sortField := strings.TrimSpace(in.sortByField); sortField != "" {
		sortOrder := strings.TrimSpace(in.sortByOrder)
		if sortOrder == "" {
			sortOrder = "DESCENDING"
		}
		query["sortBy"] = fmt.Sprintf("(field:%s,order:%s)", sortField, sortOrder)
	}

	return s.proxy.requestJSON(ctx, userID, mcpTool, "GET", "adAnalytics", query, nil, nil)
}

func (s *service) searchCreatives(ctx context.Context, userID, mcpTool string, in searchCreativesInput) (any, error) {
	accountID := strings.TrimSpace(in.accountID)
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required")
	}
	if len(in.campaignIDs) == 0 {
		return nil, fmt.Errorf("campaign_ids or campaign_urns is required")
	}

	query := map[string]string{
		"q":         "criteria",
		"campaigns": fmt.Sprintf("List(%s)", strings.Join(toCampaignURNs(in.campaignIDs), ",")),
	}
	if sortOrder := strings.TrimSpace(in.sortOrder); sortOrder != "" {
		query["sortOrder"] = sortOrder
	}
	if in.pageSize != nil && *in.pageSize > 0 {
		query["pageSize"] = strconv.Itoa(*in.pageSize)
	}
	if token := strings.TrimSpace(in.pageToken); token != "" {
		query["pageToken"] = token
	}

	headers := map[string]string{"X-RestLi-Method": "FINDER"}
	path := fmt.Sprintf("adAccounts/%s/creatives", url.PathEscape(accountID))
	return s.proxy.requestJSON(ctx, userID, mcpTool, "GET", path, query, nil, headers)
}

func parseListAdAccountsInput(params map[string]any) (listAdAccountsInput, error) {
	in := listAdAccountsInput{
		statusFilter:     toStringSlice(params["status_filter"]),
		accountIDs:       toStringSlice(params["account_ids"]),
		nameFilter:       toStringSlice(params["name_filter"]),
		referenceFilter:  toStringSlice(params["reference_filter"]),
	}
	if testValue, ok := params["test_filter"].(bool); ok {
		in.testFilter = &testValue
	}
	if count, ok := toInt(params["page_size"]); ok && count > 0 {
		in.pageSize = &count
	}
	if start, ok := toInt(params["start"]); ok && start >= 0 {
		in.start = &start
	}
	return in, nil
}

func parseGetCampaignsInput(params map[string]any) (getCampaignsInput, error) {
	accountID := strings.TrimSpace(toString(params["account_id"]))
	if accountID == "" {
		return getCampaignsInput{}, fmt.Errorf("account_id is required")
	}

	autoPaginate := true
	if autoPageBool, hasAutoPage := params["auto_paginate"].(bool); hasAutoPage {
		autoPaginate = autoPageBool
	}

	pageSize := defaultCampaignsPageSize
	if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
		pageSize = ps
	}

	pageToken := strings.TrimSpace(toString(params["page_token"]))

	var testFilter *bool
	if testValue, ok := params["test_filter"].(bool); ok {
		testFilter = &testValue
	}

	return getCampaignsInput{
		accountID:      accountID,
		statusFilter:   toStringSlice(params["status_filter"]),
		campaignGroups: toStringSlice(params["campaign_group_filter"]),
		typeFilter:     toStringSlice(params["type_filter"]),
		nameFilter:     toStringSlice(params["name_filter"]),
		testFilter:     testFilter,
		sortOrder:      strings.TrimSpace(toString(params["sort_order"])),
		autoPaginate:   autoPaginate,
		pageSize:       pageSize,
		pageToken:      pageToken,
	}, nil
}

func parseGetAnalyticsInput(params map[string]any) (getAnalyticsInput, error) {
	accountID := strings.TrimSpace(toString(params["account_id"]))
	startDate := strings.TrimSpace(toString(params["date_range_start"]))
	if accountID == "" {
		return getAnalyticsInput{}, fmt.Errorf("account_id is required")
	}
	if startDate == "" {
		return getAnalyticsInput{}, fmt.Errorf("date_range_start is required")
	}

	var endDate string
	if v := strings.TrimSpace(toString(params["date_range_end"])); v != "" {
		endDate = v
	}

	return getAnalyticsInput{
		accountID:        accountID,
		startDate:        startDate,
		endDate:          endDate,
		finderType:       toString(params["finder_type"]),
		pivots:           toStringSlice(params["pivots"]),
		timeGranularity:  toString(params["time_granularity"]),
		campaignIDs:      toStringSlice(params["campaign_ids"]),
		creativeIDs:      toStringSlice(params["creative_ids"]),
		campaignGroupIDs: toStringSlice(params["campaign_group_ids"]),
		fields:           toStringSlice(params["fields"]),
		sortByField:      toString(params["sort_by_field"]),
		sortByOrder:      toString(params["sort_by_order"]),
	}, nil
}

func parseSearchCreativesInput(params map[string]any) (searchCreativesInput, error) {
	accountID := strings.TrimSpace(toString(params["account_id"]))
	if accountID == "" {
		return searchCreativesInput{}, fmt.Errorf("account_id is required")
	}

	campaignIDs := toStringSlice(params["campaign_ids"])
	if len(campaignIDs) == 0 {
		campaignIDs = toStringSlice(params["campaign_urns"])
	}
	if len(campaignIDs) == 0 {
		return searchCreativesInput{}, fmt.Errorf("campaign_ids or campaign_urns is required")
	}

	in := searchCreativesInput{
		accountID:    accountID,
		campaignIDs:  campaignIDs,
		sortOrder:    toString(params["sort_order"]),
		pageToken:    toString(params["page_token"]),
	}
	if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
		in.pageSize = &ps
	}
	return in, nil
}

func appendListFinder(query map[string]string, field string, values []string) {
	if len(values) == 0 {
		return
	}
	encoded := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			encoded = append(encoded, trimmed)
		}
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

func toCampaignGroupURNs(items []string) []string {
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

func buildDateRangeEnd(raw string) string {
	value := strings.TrimSpace(raw)
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
	if v, ok := value.(string); ok {
		return v
	}
	return ""
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
