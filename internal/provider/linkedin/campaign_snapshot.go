package linkedin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CampaignSnapshot is a normalized campaign row for pacing and future write tools.
type CampaignSnapshot struct {
	ID               string
	URN              string
	Name             string
	Status           string
	Test             bool
	CampaignGroupURN string
	BudgetType       string // daily, lifetime, none
	BudgetAmount     float64
	CurrencyCode     string
	RunScheduleStart *int64 // Unix ms UTC when present
	RunScheduleEnd   *int64
}

const (
	budgetTypeDaily    = "daily"
	budgetTypeLifetime = "lifetime"
	budgetTypeNone     = "none"

	sponsoredCampaignPrefix      = "urn:li:sponsoredCampaign:"
	sponsoredCampaignGroupPrefix = "urn:li:sponsoredCampaignGroup:"
)

func parseCampaignSnapshotsFromElements(elements []any) []CampaignSnapshot {
	out := make([]CampaignSnapshot, 0, len(elements))
	for _, item := range elements {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if snap, ok := parseCampaignSnapshot(row); ok {
			out = append(out, snap)
		}
	}
	return out
}

func parseCampaignSnapshot(row map[string]any) (CampaignSnapshot, bool) {
	id := campaignIDFromRow(row)
	if id == "" {
		return CampaignSnapshot{}, false
	}

	snap := CampaignSnapshot{
		ID:               id,
		URN:              sponsoredCampaignPrefix + id,
		Name:             strings.TrimSpace(toString(row["name"])),
		Status:           strings.TrimSpace(toString(row["status"])),
		Test:             parseOptionalBoolParam(row["test"], false),
		CampaignGroupURN: normalizeCampaignGroupURN(toString(row["campaignGroup"])),
		BudgetType:       budgetTypeNone,
	}
	snap.RunScheduleStart, snap.RunScheduleEnd = parseRunSchedule(row["runSchedule"])

	dailyAmount, dailyCurrency, hasDaily := parseBudgetObject(row["dailyBudget"])
	totalAmount, totalCurrency, hasTotal := parseBudgetObject(row["totalBudget"])

	switch {
	case hasDaily:
		snap.BudgetType = budgetTypeDaily
		snap.BudgetAmount = dailyAmount
		snap.CurrencyCode = dailyCurrency
	case hasTotal:
		snap.BudgetType = budgetTypeLifetime
		snap.BudgetAmount = totalAmount
		snap.CurrencyCode = totalCurrency
	}

	return snap, true
}

func campaignIDFromRow(row map[string]any) string {
	if raw, ok := row["id"]; ok {
		switch v := raw.(type) {
		case string:
			return strings.TrimSpace(v)
		case float64:
			return strconv.FormatInt(int64(v), 10)
		case int:
			return strconv.Itoa(v)
		case int64:
			return strconv.FormatInt(v, 10)
		case json.Number:
			if n, err := v.Int64(); err == nil {
				return strconv.FormatInt(n, 10)
			}
		}
	}
	return ""
}

func parseBudgetObject(raw any) (amount float64, currency string, ok bool) {
	obj, okMap := raw.(map[string]any)
	if !okMap {
		return 0, "", false
	}
	amount, okAmount := budgetAmountFromField(obj["amount"])
	if !okAmount {
		return 0, "", false
	}
	return amount, strings.TrimSpace(toString(obj["currencyCode"])), true
}

func budgetAmountFromField(raw any) (float64, bool) {
	switch v := raw.(type) {
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
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return numericField(map[string]any{"v": raw}, "v")
	}
}

func parseRunSchedule(raw any) (start, end *int64) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, nil
	}
	if ms, ok := epochMsFromField(obj["start"]); ok {
		start = &ms
	}
	if ms, ok := epochMsFromField(obj["end"]); ok {
		end = &ms
	}
	return start, end
}

func epochMsFromField(raw any) (int64, bool) {
	switch v := raw.(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0, false
		}
		n, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func normalizeCampaignGroupURN(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, sponsoredCampaignGroupPrefix) {
		return trimmed
	}
	return sponsoredCampaignGroupPrefix + trimmed
}

func campaignSnapshotsFromResponse(raw any) ([]CampaignSnapshot, error) {
	pageMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected campaigns response shape")
	}
	elements, ok := pageMap["elements"].([]any)
	if !ok {
		return nil, nil
	}
	return parseCampaignSnapshotsFromElements(elements), nil
}

func filterSnapshotsByTestFlag(snapshots []CampaignSnapshot, includeTest bool) []CampaignSnapshot {
	if includeTest {
		return snapshots
	}
	filtered := make([]CampaignSnapshot, 0, len(snapshots))
	for _, snap := range snapshots {
		if !snap.Test {
			filtered = append(filtered, snap)
		}
	}
	return filtered
}

func campaignIDFromPivotValues(raw any) string {
	pivots, ok := raw.([]any)
	if !ok || len(pivots) == 0 {
		return ""
	}
	urn, ok := pivots[0].(string)
	if !ok {
		return ""
	}
	return campaignIDToNumeric(urn)
}

func filterSnapshotsByCampaignIDs(snapshots []CampaignSnapshot, campaignIDs []string) []CampaignSnapshot {
	if len(campaignIDs) == 0 {
		return snapshots
	}
	allowed := make(map[string]struct{}, len(campaignIDs))
	for _, id := range campaignIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		allowed[campaignIDToNumeric(trimmed)] = struct{}{}
	}
	if len(allowed) == 0 {
		return snapshots
	}
	filtered := make([]CampaignSnapshot, 0, len(snapshots))
	for _, snap := range snapshots {
		if _, ok := allowed[snap.ID]; ok {
			filtered = append(filtered, snap)
		}
	}
	return filtered
}
