package linkedin

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	finderTypeAnalytics         = "analytics"
	finderTypeStatistics        = "statistics"
	finderTypeAttributedRevenue = "attributedRevenueMetrics"
	maxMultiPivotCount          = 3
	minRevenueDateRangeDays     = 30
	maxRevenueDateRangeDays     = 366
	fieldRevenueAttribution     = "revenueAttributionMetrics"
)

var revenueAllowedPivots = map[string]struct{}{
	"ACCOUNT":        {},
	"CAMPAIGN_GROUP": {},
	"CAMPAIGN":       {},
}

// Nested keys inside revenueAttributionMetrics; not valid as top-level fields projections.
var revenueNestedMetrics = map[string]struct{}{
	"revenueWonInUsd":        {},
	"returnOnAdSpend":        {},
	"closedWonOpportunities": {},
	"openOpportunities":      {},
	"opportunityAmountInUsd": {},
	"opportunityWinRate":     {},
	"averageDealSizeInUsd":   {},
	"averageDaysToClose":     {},
}

func normalizeFinderType(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return finderTypeAnalytics
	}
	return trimmed
}

func hasDemographicPivot(pivots []string) bool {
	for _, pivot := range pivots {
		if strings.HasPrefix(strings.TrimSpace(pivot), "MEMBER_") {
			return true
		}
	}
	return false
}

func trimPivots(pivots []string) []string {
	trimmed := make([]string, 0, len(pivots))
	for _, pivot := range pivots {
		value := strings.TrimSpace(pivot)
		if value != "" {
			trimmed = append(trimmed, value)
		}
	}
	return trimmed
}

func applyAnalyticsPivots(query map[string]string, finderType string, pivots []string) error {
	trimmed := trimPivots(pivots)

	switch finderType {
	case finderTypeStatistics, finderTypeAttributedRevenue:
		if len(trimmed) == 0 || len(trimmed) > maxMultiPivotCount {
			return fmt.Errorf("%s finder requires 1 to %d pivots", finderType, maxMultiPivotCount)
		}
		if finderType == finderTypeAttributedRevenue {
			for _, pivot := range trimmed {
				if _, ok := revenueAllowedPivots[pivot]; !ok {
					return fmt.Errorf("attributedRevenueMetrics pivot %q is invalid; allowed: ACCOUNT, CAMPAIGN_GROUP, CAMPAIGN", pivot)
				}
			}
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

func applyAnalyticsAccountFacet(query map[string]string, finderType, accountID string) {
	urn := url.QueryEscape("urn:li:sponsoredAccount:" + accountID)
	listValue := fmt.Sprintf("List(%s)", urn)
	delete(query, "accounts")
	delete(query, "account")
	if finderType == finderTypeAttributedRevenue {
		query["account"] = listValue
		return
	}
	query["accounts"] = listValue
}

func resolveAnalyticsEndDate(endDateRaw string) time.Time {
	if end, err := parseDate(endDateRaw); err == nil {
		return end
	}
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func validateRevenueDateRange(start, end time.Time) error {
	if end.Before(start) {
		return fmt.Errorf("date_range_end must be on or after date_range_start")
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days < minRevenueDateRangeDays || days > maxRevenueDateRangeDays {
		return fmt.Errorf(
			"attributedRevenueMetrics requires date range between %d and %d days inclusive, got %d days",
			minRevenueDateRangeDays, maxRevenueDateRangeDays, days,
		)
	}
	return nil
}

func resolveDefaultAnalyticsFields(finderType string, pivots, requested []string) []string {
	if len(requested) > 0 {
		fields := requested
		if len(pivots) > 0 && !slices.Contains(fields, "pivotValues") {
			fields = append([]string{"pivotValues"}, fields...)
		}
		switch finderType {
		case finderTypeAttributedRevenue:
			return normalizeRevenueFields(fields)
		case finderTypeAnalytics:
			if hasDemographicPivot(pivots) {
				return fields
			}
			return ensureDeliveryMetricFields(fields)
		default:
			return ensureDeliveryMetricFields(fields)
		}
	}

	if len(pivots) == 0 {
		return nil
	}

	switch finderType {
	case finderTypeAttributedRevenue:
		return []string{"pivotValues", "dateRange", fieldRevenueAttribution}
	case finderTypeAnalytics:
		if hasDemographicPivot(pivots) {
			return []string{"pivotValues", fieldImpressions, "clicks", "costInLocalCurrency"}
		}
		return []string{"pivotValues", fieldApproximateMemberReach, fieldImpressions}
	default:
		return []string{"pivotValues", fieldApproximateMemberReach, fieldImpressions}
	}
}

func normalizeRevenueFields(fields []string) []string {
	nested := make([]string, 0)
	result := make([]string, 0, len(fields)+1)
	hasRevenueObject := false

	for _, field := range fields {
		trimmed := strings.TrimSpace(field)
		if trimmed == "" {
			continue
		}
		if trimmed == fieldRevenueAttribution {
			hasRevenueObject = true
			result = append(result, trimmed)
			continue
		}
		if strings.HasPrefix(trimmed, fieldRevenueAttribution+":") {
			hasRevenueObject = true
			result = append(result, trimmed)
			continue
		}
		if _, ok := revenueNestedMetrics[trimmed]; ok {
			nested = append(nested, trimmed)
			continue
		}
		result = append(result, trimmed)
	}

	if len(nested) > 0 {
		result = append(result, fmt.Sprintf("%s:(%s)", fieldRevenueAttribution, strings.Join(nested, ",")))
	} else if !hasRevenueObject && len(result) > 0 {
		result = append(result, fieldRevenueAttribution)
	}

	return result
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

	if finderType == finderTypeAttributedRevenue {
		end := resolveAnalyticsEndDate(in.endDate)
		if err := validateRevenueDateRange(start, end); err != nil {
			return nil, err
		}
	}

	query := map[string]string{
		"q": finderType,
		"dateRange": fmt.Sprintf(
			"(start:(year:%d,month:%d,day:%d)%s)",
			start.Year(), int(start.Month()), start.Day(), buildDateRangeEnd(in.endDate),
		),
	}
	applyAnalyticsAccountFacet(query, finderType, accountID)

	if err := applyAnalyticsPivots(query, finderType, in.pivots); err != nil {
		return nil, err
	}

	if finderType != finderTypeAttributedRevenue {
		if granularity := strings.TrimSpace(in.timeGranularity); granularity != "" {
			query["timeGranularity"] = granularity
		}
	}

	if fields := resolveDefaultAnalyticsFields(finderType, in.pivots, in.fields); len(fields) > 0 {
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

	if finderType != finderTypeAttributedRevenue {
		if sortField := strings.TrimSpace(in.sortByField); sortField != "" {
			sortOrder := strings.TrimSpace(in.sortByOrder)
			if sortOrder == "" {
				sortOrder = "DESCENDING"
			}
			query["sortBy"] = fmt.Sprintf("(field:%s,order:%s)", sortField, sortOrder)
		}
	}

	return query, nil
}
