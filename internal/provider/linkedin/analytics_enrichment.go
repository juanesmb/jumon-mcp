package linkedin

import (
	"encoding/json"
	"math"
	"slices"
	"strconv"
	"strings"
)

const (
	fieldApproximateMemberReach = "approximateMemberReach"
	fieldApproximateUniqueReach = "approximateUniqueImpressions"
	fieldImpressions            = "impressions"
	fieldAverageFrequency       = "averageFrequency"
)

// enrichAnalyticsResponse adds derived delivery metrics and pivot labels to each element.
// averageFrequency matches Campaign Manager: impressions / reach (unique members).
func enrichAnalyticsResponse(payload any, pivots []string) any {
	root, ok := payload.(map[string]any)
	if !ok {
		return payload
	}

	elements, ok := root["elements"].([]any)
	if !ok {
		return payload
	}

	for i, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		enrichAnalyticsElement(row)
		enrichPivotLabels(row, pivots)
		elements[i] = row
	}
	root["elements"] = elements
	return root
}

func enrichAnalyticsElement(row map[string]any) {
	impressions, okImpressions := numericField(row, fieldImpressions)
	reach, okReach := memberReach(row)
	if !okImpressions || !okReach || reach <= 0 {
		return
	}
	row[fieldAverageFrequency] = deriveAverageFrequency(impressions, reach)
}

func deriveAverageFrequency(impressions, reach float64) float64 {
	if reach <= 0 {
		return 0
	}
	return math.Round((impressions/reach)*100) / 100
}

func memberReach(row map[string]any) (float64, bool) {
	if reach, ok := numericField(row, fieldApproximateMemberReach); ok {
		return reach, true
	}
	return numericField(row, fieldApproximateUniqueReach)
}

func numericField(row map[string]any, key string) (float64, bool) {
	raw, ok := row[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

// ensureDeliveryMetricFields adds LinkedIn fields required to compute averageFrequency.
func ensureDeliveryMetricFields(fields []string) []string {
	for _, name := range []string{fieldApproximateMemberReach, fieldImpressions} {
		if !slices.Contains(fields, name) {
			fields = append(fields, name)
		}
	}
	return fields
}
