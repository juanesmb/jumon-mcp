package meta

import (
	"encoding/json"
	"fmt"
	"strings"
)

func normalizeActID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("meta: act_id is required")
	}
	if strings.HasPrefix(id, AdAccountIDPrefix) {
		return id, nil
	}
	if strings.Contains(id, "/") || strings.Contains(id, " ") {
		return "", fmt.Errorf("meta: act_id must be numeric or prefixed with %s", AdAccountIDPrefix)
	}
	return AdAccountIDPrefix + id, nil
}

func requireCampaignID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("meta: campaign_id is required")
	}
	return id, nil
}

type listPaginationInput struct {
	limit         int
	after         string
	before        string
	autoPaginate  bool
	fields        []string
	effectiveStatus []string
}

type listCampaignsInput struct {
	listPaginationInput
}

type listAdSetsInput struct {
	listPaginationInput
	campaignID string
}

type listAdsInput struct {
	listPaginationInput
	campaignID string
	adSetID    string
}

type timeRangeInput struct {
	Since string `json:"since"`
	Until string `json:"until"`
}

type insightsInput struct {
	fields                      []string
	datePreset                  string
	timeRange                   *timeRangeInput
	timeRanges                  []timeRangeInput
	timeIncrement               string
	level                       string
	actionAttributionWindows    []string
	actionBreakdowns            []string
	actionReportTime            string
	breakdowns                  []string
	defaultSummary              *bool
	useAccountAttributionSetting *bool
	useUnifiedAttributionSetting *bool
	filtering                   []map[string]any
	sort                        string
	since                       string
	until                       string
	limit                       int
	after                       string
	autoPaginate                bool
}

type getAdAccountInput struct {
	actID  string
	fields []string
}

type getCampaignInput struct {
	campaignID string
	fields     []string
}

type getAdSetInput struct {
	adSetID string
	fields  []string
}

type getAdInput struct {
	adID   string
	fields []string
}

type deliveryErrorsInput struct {
	entityIDs []string
}

type searchAdEntitiesInput struct {
	actID string
	insightsInput
}

type fieldContextInput struct {
	fieldNames []string
	level      string
}

type listAccountPagesInput struct {
	listPaginationInput
	actID string
}

func parseListAccountPagesInput(params map[string]any) listAccountPagesInput {
	return listAccountPagesInput{
		listPaginationInput: parseListPagination(params, defaultAccountPageFields),
		actID:               strings.TrimSpace(toString(params["act_id"])),
	}
}

func parseListPagination(params map[string]any, defaultFields []string) listPaginationInput {
	in := listPaginationInput{
		after:        strings.TrimSpace(toString(params["after"])),
		before:       strings.TrimSpace(toString(params["before"])),
		autoPaginate: parseAutoPaginate(params["auto_paginate"]),
		fields:       toStringSlice(params["fields"]),
		effectiveStatus: toStringSlice(params["effective_status"]),
	}
	if in.fields == nil {
		in.fields = append([]string(nil), defaultFields...)
	}
	if limit, ok := toInt(params["limit"]); ok && limit > 0 {
		in.limit = clampLimit(limit)
	} else {
		in.limit = defaultListLimit
	}
	return in
}

func parseInsightsInput(params map[string]any, defaultFields []string) (insightsInput, error) {
	in := insightsInput{
		fields:         toStringSlice(params["fields"]),
		datePreset:     strings.TrimSpace(toString(params["date_preset"])),
		timeIncrement:  strings.TrimSpace(toString(params["time_increment"])),
		level:          strings.TrimSpace(toString(params["level"])),
		actionReportTime: strings.TrimSpace(toString(params["action_report_time"])),
		sort:           strings.TrimSpace(toString(params["sort"])),
		since:          strings.TrimSpace(toString(params["since"])),
		until:          strings.TrimSpace(toString(params["until"])),
		after:          strings.TrimSpace(toString(params["after"])),
		autoPaginate:   parseAutoPaginate(params["auto_paginate"]),
		actionAttributionWindows: toStringSlice(params["action_attribution_windows"]),
		actionBreakdowns:         toStringSlice(params["action_breakdowns"]),
		breakdowns:               toStringSlice(params["breakdowns"]),
	}
	if in.fields == nil {
		in.fields = append([]string(nil), defaultFields...)
	}
	if in.datePreset == "" {
		in.datePreset = defaultDatePreset
	}
	if limit, ok := toInt(params["limit"]); ok && limit > 0 {
		in.limit = clampInsightsLimit(limit)
	} else {
		in.limit = defaultListLimit
	}
	if tr, err := parseTimeRange(params["time_range"]); err != nil {
		return insightsInput{}, err
	} else if tr != nil {
		in.timeRange = tr
	}
	if trs, err := parseTimeRanges(params["time_ranges"]); err != nil {
		return insightsInput{}, err
	} else if len(trs) > 0 {
		in.timeRanges = trs
	}
	if filtering, err := parseFiltering(params["filtering"]); err != nil {
		return insightsInput{}, err
	} else {
		in.filtering = filtering
	}
	in.defaultSummary = parseOptionalBool(params["default_summary"])
	in.useAccountAttributionSetting = parseOptionalBool(params["use_account_attribution_setting"])
	in.useUnifiedAttributionSetting = parseOptionalBool(params["use_unified_attribution_setting"])
	return in, nil
}

func parseAutoPaginate(value any) bool {
	if value == nil {
		return true
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return true
}

func parseOptionalBool(value any) *bool {
	if value == nil {
		return nil
	}
	if b, ok := value.(bool); ok {
		return &b
	}
	return nil
}

func parseTimeRange(value any) (*timeRangeInput, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case map[string]any:
		tr := timeRangeInput{
			Since: strings.TrimSpace(toString(v["since"])),
			Until: strings.TrimSpace(toString(v["until"])),
		}
		if tr.Since == "" && tr.Until == "" {
			return nil, nil
		}
		return &tr, nil
	default:
		return nil, fmt.Errorf("meta: time_range must be an object with since and until")
	}
}

func parseTimeRanges(value any) ([]timeRangeInput, error) {
	if value == nil {
		return nil, nil
	}
	raw, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("meta: time_ranges must be an array of {since, until} objects")
	}
	out := make([]timeRangeInput, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("meta: time_ranges entries must be objects")
		}
		tr := timeRangeInput{
			Since: strings.TrimSpace(toString(m["since"])),
			Until: strings.TrimSpace(toString(m["until"])),
		}
		out = append(out, tr)
	}
	return out, nil
}

func parseFiltering(value any) ([]map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	raw, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("meta: filtering must be an array of filter objects")
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("meta: filtering entries must be objects")
		}
		out = append(out, m)
	}
	return out, nil
}

func clampLimit(limit int) int {
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func clampInsightsLimit(limit int) int {
	if limit > maxInsightsLimit {
		return maxInsightsLimit
	}
	return limit
}

func requireAdSetID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("meta: adset_id is required")
	}
	return id, nil
}

func requireAdID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("meta: ad_id is required")
	}
	return id, nil
}

func parseDeliveryErrorsInput(params map[string]any) (deliveryErrorsInput, error) {
	ids := toStringSlice(params["entity_ids"])
	if len(ids) == 0 {
		return deliveryErrorsInput{}, fmt.Errorf("meta: entity_ids is required")
	}
	if len(ids) > maxDeliveryEntityIDs {
		return deliveryErrorsInput{}, fmt.Errorf("meta: entity_ids exceeds max %d", maxDeliveryEntityIDs)
	}
	return deliveryErrorsInput{entityIDs: ids}, nil
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			f, err := v.Float64()
			if err != nil {
				return 0, false
			}
			return int(f), true
		}
		return int(i), true
	default:
		return 0, false
	}
}

func toStringSlice(value any) []string {
	if value == nil {
		return nil
	}
	raw, ok := value.([]any)
	if !ok {
		if s := strings.TrimSpace(toString(value)); s != "" {
			return []string{s}
		}
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s := strings.TrimSpace(toString(item)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func unmarshalPayload(raw []byte) (any, error) {
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type activitiesInput struct {
	fields    []string
	timeRange *timeRangeInput
	since     string
	until     string
	limit     int
	after     string
	before    string
}

type listCustomConversionsInput struct {
	listPaginationInput
	datasetID string
}

type getDatasetInput struct {
	datasetID string
	fields    []string
}

type datasetStatsInput struct {
	datasetID   string
	startTime   string
	endTime     string
	eventName   string
	eventSource string
	aggregation string
}

type datasetQualityInput struct {
	datasetID string
	fields    string
}

type listCreativeAdsInput struct {
	listPaginationInput
	creativeID string
}

type interestSuggestionsInput struct {
	interestList []string
	limit        int
}

func requireDatasetID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("meta: dataset_id is required")
	}
	return id, nil
}

func requireCreativeID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("meta: creative_id is required")
	}
	return id, nil
}

func parseActivitiesInput(params map[string]any) (activitiesInput, error) {
	in := activitiesInput{
		fields: toStringSlice(params["fields"]),
		since:  strings.TrimSpace(toString(params["since"])),
		until:  strings.TrimSpace(toString(params["until"])),
		after:  strings.TrimSpace(toString(params["after"])),
		before: strings.TrimSpace(toString(params["before"])),
	}
	if in.fields == nil {
		in.fields = append([]string(nil), defaultActivityFields...)
	}
	if limit, ok := toInt(params["limit"]); ok && limit > 0 {
		in.limit = clampLimit(limit)
	} else {
		in.limit = defaultListLimit
	}
	if tr, err := parseTimeRange(params["time_range"]); err != nil {
		return activitiesInput{}, err
	} else if tr != nil {
		in.timeRange = tr
	}
	return in, nil
}

func parseListCustomConversionsInput(params map[string]any) listCustomConversionsInput {
	return listCustomConversionsInput{
		listPaginationInput: parseListPagination(params, defaultCustomConversionListFields),
		datasetID:           strings.TrimSpace(toString(params["dataset_id"])),
	}
}

func parseGetDatasetInput(params map[string]any) getDatasetInput {
	return getDatasetInput{
		datasetID: strings.TrimSpace(toString(params["dataset_id"])),
		fields:    toStringSlice(params["fields"]),
	}
}

func parseDatasetStatsInput(params map[string]any) (datasetStatsInput, error) {
	datasetID, err := requireDatasetID(toString(params["dataset_id"]))
	if err != nil {
		return datasetStatsInput{}, err
	}
	return datasetStatsInput{
		datasetID:   datasetID,
		startTime:   strings.TrimSpace(toString(params["start_time"])),
		endTime:     strings.TrimSpace(toString(params["end_time"])),
		eventName:   strings.TrimSpace(toString(params["event_name"])),
		eventSource: strings.TrimSpace(toString(params["event_source"])),
		aggregation: strings.TrimSpace(toString(params["aggregation"])),
	}, nil
}

func parseDatasetQualityInput(params map[string]any) (datasetQualityInput, error) {
	datasetID, err := requireDatasetID(toString(params["dataset_id"]))
	if err != nil {
		return datasetQualityInput{}, err
	}
	in := datasetQualityInput{datasetID: datasetID}
	if fields := toStringSlice(params["fields"]); len(fields) > 0 {
		in.fields = strings.Join(fields, ",")
	} else if s := strings.TrimSpace(toString(params["fields"])); s != "" {
		in.fields = s
	}
	return in, nil
}

func parseListCreativeAdsInput(params map[string]any) (listCreativeAdsInput, error) {
	creativeID, err := requireCreativeID(toString(params["creative_id"]))
	if err != nil {
		return listCreativeAdsInput{}, err
	}
	return listCreativeAdsInput{
		listPaginationInput: parseListPagination(params, defaultCreativeAdListFields),
		creativeID:          creativeID,
	}, nil
}

func parseDemographicClass(params map[string]any) string {
	class := strings.TrimSpace(toString(params["class"]))
	if class == "" {
		return "demographics"
	}
	return class
}

func parseInterestList(params map[string]any) ([]string, error) {
	list := toStringSlice(params["interest_list"])
	if len(list) == 0 {
		return nil, fmt.Errorf("meta: interest_list is required and must contain at least one item")
	}
	return list, nil
}

func parseInterestSuggestionsInput(params map[string]any) (interestSuggestionsInput, error) {
	list, err := parseInterestList(params)
	if err != nil {
		return interestSuggestionsInput{}, err
	}
	in := interestSuggestionsInput{interestList: list}
	if limit, ok := toInt(params["limit"]); ok && limit > 0 {
		in.limit = clampLimit(limit)
	} else {
		in.limit = defaultListLimit
	}
	return in, nil
}

func parseSearchLimit(params map[string]any) int {
	if limit, ok := toInt(params["limit"]); ok && limit > 0 {
		return clampLimit(limit)
	}
	return defaultListLimit
}
