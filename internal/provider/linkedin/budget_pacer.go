package linkedin

import (
	"fmt"
	"strings"
	"time"
)

// BudgetPacerInput is the parsed MCP tool input.
type BudgetPacerInput struct {
	AccountID            string
	PeriodStart          string
	PeriodEnd            string
	StatusFilter         []string
	CampaignGroupIDs     []string
	CampaignIDs          []string
	IncludeTestCampaigns bool
	AutoPaginate         bool
	PacingThresholds     PacingThresholds
	ComparePeriodStart   string
	ComparePeriodEnd     string
}

// DatePeriod is an inclusive YYYY-MM-DD range.
type DatePeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// PacerReportMetadata describes fetch coverage and warnings.
type PacerReportMetadata struct {
	Truncated         bool     `json:"truncated"`
	CampaignsIncluded int      `json:"campaignsIncluded"`
	Warnings          []string `json:"warnings"`
}

// AccountPacerSummary rolls up account-level pacing when a single currency applies.
type AccountPacerSummary struct {
	SpendToDate         float64  `json:"spendToDate"`
	ExpectedSpendToDate *float64 `json:"expectedSpendToDate,omitempty"`
	PacingPercent       *float64 `json:"pacingPercent,omitempty"`
	PacingStatus        string   `json:"pacingStatus"`
	CurrencyCode        string   `json:"currencyCode,omitempty"`
}

// PacerRow is one campaign pacing line.
type PacerRow struct {
	CampaignID           string   `json:"campaignId"`
	CampaignURN          string   `json:"campaignUrn"`
	Name                 string   `json:"name"`
	Status               string   `json:"status"`
	CampaignGroupURN     string   `json:"campaignGroupUrn,omitempty"`
	BudgetType           string   `json:"budgetType"`
	BudgetAmount         float64  `json:"budgetAmount,omitempty"`
	CurrencyCode         string   `json:"currencyCode,omitempty"`
	SpendToDate          float64  `json:"spendToDate"`
	ExpectedSpendToDate  *float64 `json:"expectedSpendToDate,omitempty"`
	PacingPercent        *float64 `json:"pacingPercent,omitempty"`
	PacingStatus         string   `json:"pacingStatus"`
	ProjectedPeriodSpend *float64 `json:"projectedPeriodSpend,omitempty"`
	SpendPrior           *float64 `json:"spendPrior,omitempty"`
	SpendChangePercent   *float64 `json:"spendChangePercent,omitempty"`
}

// CampaignGroupPacerRollup aggregates rows by campaign group.
type CampaignGroupPacerRollup struct {
	CampaignGroupURN    string   `json:"campaignGroupUrn"`
	CampaignGroupName   string   `json:"campaignGroupName,omitempty"`
	SpendToDate         float64  `json:"spendToDate"`
	ExpectedSpendToDate *float64 `json:"expectedSpendToDate,omitempty"`
	PacingPercent       *float64 `json:"pacingPercent,omitempty"`
	PacingStatus        string   `json:"pacingStatus"`
	CampaignCount       int      `json:"campaignCount"`
	CurrencyCode        string   `json:"currencyCode,omitempty"`
}

// BudgetPacerReport is the MCP tool response contract.
type BudgetPacerReport struct {
	AccountID       string                     `json:"accountId"`
	Period          DatePeriod                 `json:"period"`
	ComparePeriod   *DatePeriod                `json:"comparePeriod,omitempty"`
	Metadata        PacerReportMetadata        `json:"metadata"`
	Summary         *AccountPacerSummary       `json:"summary,omitempty"`
	Rows            []PacerRow                 `json:"rows"`
	ByCampaignGroup []CampaignGroupPacerRollup `json:"byCampaignGroup"`
}

func parseBudgetPacerInput(params map[string]any) (BudgetPacerInput, error) {
	accountID := strings.TrimSpace(toString(params["account_id"]))
	periodStart := strings.TrimSpace(toString(params["date_range_start"]))
	if accountID == "" {
		return BudgetPacerInput{}, fmt.Errorf("account_id is required")
	}
	if periodStart == "" {
		return BudgetPacerInput{}, fmt.Errorf("date_range_start is required")
	}

	statusFilter := toStringSlice(params["status_filter"])
	if len(statusFilter) == 0 {
		statusFilter = []string{"ACTIVE"}
	}

	in := BudgetPacerInput{
		AccountID:            accountID,
		PeriodStart:          periodStart,
		PeriodEnd:            strings.TrimSpace(toString(params["date_range_end"])),
		StatusFilter:         statusFilter,
		CampaignGroupIDs:     toStringSlice(params["campaign_group_ids"]),
		CampaignIDs:          toStringSlice(params["campaign_ids"]),
		IncludeTestCampaigns: parseOptionalBoolParam(params["include_test_campaigns"], false),
		AutoPaginate:         parseOptionalBoolParam(params["auto_paginate"], true),
		ComparePeriodStart:   strings.TrimSpace(toString(params["compare_date_range_start"])),
		ComparePeriodEnd:     strings.TrimSpace(toString(params["compare_date_range_end"])),
		PacingThresholds:     parsePacingThresholds(params["pacing_thresholds"]),
	}
	return in, nil
}

func parsePacingThresholds(raw any) PacingThresholds {
	obj, ok := raw.(map[string]any)
	if !ok {
		return defaultPacingThresholds()
	}
	out := defaultPacingThresholds()
	if over, ok := budgetAmountFromField(obj["over"]); ok && over > 0 {
		out.Over = over
	}
	if under, ok := budgetAmountFromField(obj["under"]); ok && under > 0 {
		out.Under = under
	}
	return out
}

func resolvePeriodDates(startRaw, endRaw string) (time.Time, time.Time, error) {
	start, err := parseDate(startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end := resolveAnalyticsEndDate(endRaw)
	return truncateUTCDate(start), truncateUTCDate(end), nil
}

func linkedInBudgetPacerToolDescription() string {
	return "Returns a budget pacing report for one ad account: per-campaign spend vs expected spend, pacing percent and status, optional campaign-group rollups and account summary. " +
		"Combines campaign budget config (linkedin_get_campaigns data) with period spend (adAnalytics). " +
		"Prefer this tool for pacing questions instead of manually joining linkedin_get_campaigns and linkedin_get_ad_analytics. " +
		"Use linkedin_get_ad_analytics directly for custom metrics, demographics, or conversion pivots. " +
		"Optional compare_date_range_* adds prior-period spend and spendChangePercent per row. " +
		"Pacing is computed in UTC calendar days; projectedPeriodSpend uses linear extrapolation (not LinkedIn delivery pacing)."
}

func budgetPacerReportSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id", "date_range_start"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID.",
			},
			"date_range_start": map[string]any{
				"type":        "string",
				"description": "Inclusive pacing period start (YYYY-MM-DD, UTC).",
			},
			"date_range_end": map[string]any{
				"type":        "string",
				"description": "Inclusive pacing period end (YYYY-MM-DD, UTC). Defaults to today when omitted.",
			},
			"status_filter": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Campaign statuses to include. Default [ACTIVE].",
			},
			"campaign_group_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Numeric campaign group IDs or URNs to scope the report.",
			},
			"campaign_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Numeric campaign IDs or URNs to scope the report.",
			},
			"include_test_campaigns": map[string]any{
				"type":        "boolean",
				"description": "When false (default), excludes test campaigns.",
			},
			"auto_paginate": map[string]any{
				"type":        "boolean",
				"description": "When true (default), fetches all campaign pages up to the server cap.",
			},
			"pacing_thresholds": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"over":  map[string]any{"type": "number", "description": "Ratio above which pacingStatus is over (default 1.1)."},
					"under": map[string]any{"type": "number", "description": "Ratio below which pacingStatus is under (default 0.9)."},
				},
			},
			"compare_date_range_start": map[string]any{
				"type":        "string",
				"description": "Optional prior period start for spendPrior and spendChangePercent.",
			},
			"compare_date_range_end": map[string]any{
				"type":        "string",
				"description": "Optional prior period end (YYYY-MM-DD). Defaults to today when compare start set and end omitted.",
			},
		},
	}
}
