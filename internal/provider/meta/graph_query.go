package meta

import (
	"encoding/json"
	"strings"
)

func joinCSV(values []string) string {
	return strings.Join(values, ",")
}

func jsonEncode(value any) string {
	if value == nil {
		return ""
	}
	b, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(b)
}

func buildListQuery(in listPaginationInput) map[string]string {
	q := map[string]string{
		"fields": joinCSV(in.fields),
		"limit":  toString(in.limit),
	}
	if in.after != "" {
		q["after"] = in.after
	}
	if in.before != "" {
		q["before"] = in.before
	}
	if len(in.effectiveStatus) > 0 {
		q["effective_status"] = jsonEncode(in.effectiveStatus)
	}
	return q
}

func buildInsightsQuery(in insightsInput) map[string]string {
	q := map[string]string{
		"fields": joinCSV(in.fields),
		"limit":  toString(in.limit),
	}
	if in.after != "" {
		q["after"] = in.after
	}
	if in.level != "" {
		q["level"] = in.level
	}
	if len(in.breakdowns) > 0 {
		q["breakdowns"] = joinCSV(in.breakdowns)
	}
	if len(in.actionAttributionWindows) > 0 {
		q["action_attribution_windows"] = joinCSV(in.actionAttributionWindows)
	}
	if len(in.actionBreakdowns) > 0 {
		q["action_breakdowns"] = joinCSV(in.actionBreakdowns)
	}
	if in.actionReportTime != "" {
		q["action_report_time"] = in.actionReportTime
	}
	if in.timeIncrement != "" && in.timeIncrement != "all_days" {
		q["time_increment"] = in.timeIncrement
	}
	if in.sort != "" {
		q["sort"] = in.sort
	}
	if len(in.filtering) > 0 {
		q["filtering"] = jsonEncode(in.filtering)
	}
	if in.defaultSummary != nil && *in.defaultSummary {
		q["default_summary"] = "true"
	}
	if in.useAccountAttributionSetting != nil && *in.useAccountAttributionSetting {
		q["use_account_attribution_setting"] = "true"
	}
	if in.useUnifiedAttributionSetting != nil && *in.useUnifiedAttributionSetting {
		q["use_unified_attribution_setting"] = "true"
	} else if in.useUnifiedAttributionSetting == nil {
		q["use_unified_attribution_setting"] = "true"
	}

	applyInsightsTimeParams(q, in)
	return q
}

func applyInsightsTimeParams(q map[string]string, in insightsInput) {
	if len(in.timeRanges) > 0 {
		q["time_ranges"] = jsonEncode(in.timeRanges)
		return
	}
	if in.timeRange != nil {
		q["time_range"] = jsonEncode(in.timeRange)
		return
	}
	if in.since != "" || in.until != "" {
		if in.since != "" {
			q["since"] = in.since
		}
		if in.until != "" {
			q["until"] = in.until
		}
		return
	}
	if in.datePreset != "" {
		q["date_preset"] = in.datePreset
	}
}

func buildActivitiesQuery(in activitiesInput) map[string]string {
	q := map[string]string{
		"fields": joinCSV(in.fields),
		"limit":  toString(in.limit),
	}
	if in.after != "" {
		q["after"] = in.after
	}
	if in.before != "" {
		q["before"] = in.before
	}
	if in.timeRange != nil {
		q["time_range"] = jsonEncode(in.timeRange)
	} else {
		if in.since != "" {
			q["since"] = in.since
		}
		if in.until != "" {
			q["until"] = in.until
		}
	}
	return q
}

func buildDatasetStatsQuery(in datasetStatsInput) map[string]string {
	q := map[string]string{}
	if in.startTime != "" {
		q["start_time"] = in.startTime
	}
	if in.endTime != "" {
		q["end_time"] = in.endTime
	}
	if in.eventName != "" {
		q["event_name"] = in.eventName
	}
	if in.eventSource != "" {
		q["event_source"] = in.eventSource
	}
	if in.aggregation != "" {
		q["aggregation"] = in.aggregation
	}
	return q
}

func buildDatasetQualityQuery(in datasetQualityInput) map[string]string {
	q := map[string]string{
		"dataset_id": in.datasetID,
	}
	fields := in.fields
	if fields == "" {
		fields = defaultDatasetQualityFields
	}
	q["fields"] = fields
	return q
}
