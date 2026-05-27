package googleads

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultReportLimit = 500
	maxReportLimit     = 1000
)

func googleBuildWhereClause(parts []string) string {
	filters := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			filters = append(filters, trimmed)
		}
	}
	if len(filters) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(filters, " AND ")
}

func googleInClause(field string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	clean := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	if len(clean) == 0 {
		return ""
	}
	return fmt.Sprintf("%s IN (%s)", field, strings.Join(clean, ","))
}

func googleQuotedInClause(field string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	clean := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		escaped := strings.ReplaceAll(trimmed, "'", "\\'")
		clean = append(clean, fmt.Sprintf("'%s'", escaped))
	}
	if len(clean) == 0 {
		return ""
	}
	return fmt.Sprintf("%s IN (%s)", field, strings.Join(clean, ","))
}

func googleEnumInClause(field string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	clean := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized != "" {
			clean = append(clean, normalized)
		}
	}
	if len(clean) == 0 {
		return ""
	}
	return fmt.Sprintf("%s IN (%s)", field, strings.Join(clean, ","))
}

func googleLikeClause(field, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	escaped := strings.ReplaceAll(trimmed, "'", "\\'")
	return fmt.Sprintf("%s LIKE '%%%s%%'", field, escaped)
}

func googleExactClause(field, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	escaped := strings.ReplaceAll(trimmed, "'", "\\'")
	return fmt.Sprintf("%s = '%s'", field, escaped)
}

func googleDateBetweenClause(start, end string) string {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" && end == "" {
		return ""
	}
	if start == "" {
		return fmt.Sprintf("segments.date <= '%s'", end)
	}
	if end == "" {
		return fmt.Sprintf("segments.date >= '%s'", start)
	}
	return fmt.Sprintf("segments.date BETWEEN '%s' AND '%s'", start, end)
}

func googleLimitClause(limit int) string {
	if limit <= 0 {
		return ""
	}
	return fmt.Sprintf(" LIMIT %d", limit)
}

func googleConversionActionResourceInClause(customerID string, ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	resources := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "/") {
			resources = append(resources, trimmed)
			continue
		}
		resources = append(resources, fmt.Sprintf("customers/%s/conversionActions/%s", customerID, trimmed))
	}
	return googleQuotedInClause("segments.conversion_action", resources)
}

func normalizeReportLimit(raw int) int {
	if raw <= 0 {
		return defaultReportLimit
	}
	if raw > maxReportLimit {
		return maxReportLimit
	}
	return raw
}

func parseLimitParam(params map[string]any) int {
	switch v := params["limit"].(type) {
	case float64:
		return normalizeReportLimit(int(v))
	case int:
		return normalizeReportLimit(v)
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return normalizeReportLimit(n)
		}
	}
	return defaultReportLimit
}
