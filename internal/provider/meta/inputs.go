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

type searchAdEntitiesInput struct {
	actID string
	insightsInput
}

type fieldContextInput struct {
	fieldNames []string
	level      string
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
		in.limit = clampLimit(limit)
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
