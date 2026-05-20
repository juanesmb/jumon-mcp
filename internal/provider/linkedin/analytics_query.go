package linkedin

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

const (
	finderTypeAnalytics         = "analytics"
	finderTypeStatistics        = "statistics"
	maxStatisticsPivots         = 3
)

func normalizeFinderType(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return finderTypeAnalytics
	}
	return trimmed
}

func applyAnalyticsPivots(query map[string]string, finderType string, pivots []string) error {
	trimmed := make([]string, 0, len(pivots))
	for _, pivot := range pivots {
		value := strings.TrimSpace(pivot)
		if value != "" {
			trimmed = append(trimmed, value)
		}
	}

	switch finderType {
	case finderTypeStatistics:
		if len(trimmed) == 0 || len(trimmed) > maxStatisticsPivots {
			return fmt.Errorf("statistics finder requires 1 to %d pivots", maxStatisticsPivots)
		}
		query["pivots"] = fmt.Sprintf("List(%s)", strings.Join(trimmed, ","))
		return nil
	default:
		if len(trimmed) == 0 {
			return fmt.Errorf("at least one pivot is required")
		}
		query["pivot"] = trimmed[0]
		return nil
	}
}

func buildAdAnalyticsQuery(in getAnalyticsInput) (map[string]string, error) {
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

	finderType := normalizeFinderType(in.finderType)

	query := map[string]string{
		"q":        finderType,
		"accounts": fmt.Sprintf("List(%s)", url.QueryEscape("urn:li:sponsoredAccount:"+accountID)),
		"dateRange": fmt.Sprintf(
			"(start:(year:%d,month:%d,day:%d)%s)",
			start.Year(), int(start.Month()), start.Day(), buildDateRangeEnd(in.endDate),
		),
	}

	if err := applyAnalyticsPivots(query, finderType, in.pivots); err != nil {
		return nil, err
	}

	if granularity := strings.TrimSpace(in.timeGranularity); granularity != "" {
		query["timeGranularity"] = granularity
	}

	fields := in.fields
	if len(in.pivots) > 0 {
		if len(fields) == 0 {
			fields = []string{"pivotValues", fieldApproximateMemberReach, fieldImpressions}
		} else {
			if !slices.Contains(fields, "pivotValues") {
				fields = append([]string{"pivotValues"}, fields...)
			}
			fields = ensureDeliveryMetricFields(fields)
		}
	}
	if len(fields) > 0 {
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

	return query, nil
}
