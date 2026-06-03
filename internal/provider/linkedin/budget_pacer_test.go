package linkedin

import (
	"strings"
	"testing"
	"time"
)

func TestParseBudgetPacerInput_defaults(t *testing.T) {
	t.Parallel()

	in, err := parseBudgetPacerInput(map[string]any{
		"account_id":       "512247261",
		"date_range_start": "2026-06-01",
		"date_range_end":   "2026-06-10",
	})
	if err != nil {
		t.Fatalf("parseBudgetPacerInput() error = %v", err)
	}
	if len(in.StatusFilter) != 1 || in.StatusFilter[0] != "ACTIVE" {
		t.Fatalf("statusFilter = %v", in.StatusFilter)
	}
	if in.AutoPaginate != true {
		t.Fatal("expected auto paginate default true")
	}
	if in.PacingThresholds.Over != 1.1 {
		t.Fatalf("over threshold = %v", in.PacingThresholds.Over)
	}
}

func TestBuildPacerRows_andGroupRollups(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	snapshots := []CampaignSnapshot{
		{
			ID: "1", URN: "urn:li:sponsoredCampaign:1", Name: "A", Status: "ACTIVE",
			CampaignGroupURN: "urn:li:sponsoredCampaignGroup:10",
			BudgetType:       budgetTypeDaily, BudgetAmount: 100, CurrencyCode: "USD",
		},
		{
			ID: "2", URN: "urn:li:sponsoredCampaign:2", Name: "B", Status: "ACTIVE",
			CampaignGroupURN: "urn:li:sponsoredCampaignGroup:10",
			BudgetType:       budgetTypeDaily, BudgetAmount: 50, CurrencyCode: "USD",
		},
	}
	spend := map[string]float64{"1": 1500, "2": 200}
	rows := buildPacerRows(snapshots, spend, nil, start, end, defaultPacingThresholds(), nil)
	if len(rows) != 2 {
		t.Fatalf("rows = %d", len(rows))
	}
	rollups := buildGroupRollups(rows, map[string]string{
		"urn:li:sponsoredCampaignGroup:10": "MOFU",
	}, defaultPacingThresholds())
	if len(rollups) != 1 {
		t.Fatalf("rollups = %+v", rollups)
	}
	if rollups[0].CampaignGroupName != "MOFU" || rollups[0].CampaignCount != 2 {
		t.Fatalf("rollup = %+v", rollups[0])
	}
}

func TestParseSpendByCampaignID_stringCost(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"elements": []any{
			map[string]any{
				"pivotValues":         []any{"urn:li:sponsoredCampaign:394073893"},
				"costInLocalCurrency": "51.638580226479322425",
			},
		},
	}
	spend := parseSpendByCampaignID(raw)
	if got := spend["394073893"]; got < 51.6 || got > 51.7 {
		t.Fatalf("spend[394073893] = %v", got)
	}
}

func TestBuildAccountSummary_mixedCurrencyReturnsNil(t *testing.T) {
	t.Parallel()

	rows := []PacerRow{
		{CurrencyCode: "USD", SpendToDate: 100, ExpectedSpendToDate: floatPtr(90)},
		{CurrencyCode: "EUR", SpendToDate: 50, ExpectedSpendToDate: floatPtr(40)},
	}
	if buildAccountSummary(rows, defaultPacingThresholds(), nil) != nil {
		t.Fatal("expected nil summary for mixed currencies")
	}
}

func TestParseBudgetPacerInput_startDateAlias(t *testing.T) {
	t.Parallel()

	in, err := parseBudgetPacerInput(map[string]any{
		"account_id": "512247261",
		"start_date": "2026-06-01",
		"end_date":   "2026-06-03",
	})
	if err != nil {
		t.Fatalf("parseBudgetPacerInput() error = %v", err)
	}
	if in.PeriodStart != "2026-06-01" || in.PeriodEnd != "2026-06-03" {
		t.Fatalf("period = %q..%q", in.PeriodStart, in.PeriodEnd)
	}
}

func TestParseBudgetPacerInput_missingDateHintsParams(t *testing.T) {
	t.Parallel()

	_, err := parseBudgetPacerInput(map[string]any{
		"account_id":  "512247261",
		"date_start":  "2026-06-01",
	})
	if err == nil {
		t.Fatal("expected error without date_range_start")
	}
	msg := err.Error()
	if !strings.Contains(msg, "date_range_start") || !strings.Contains(msg, "account_id") {
		t.Fatalf("error = %q", msg)
	}
	if !strings.Contains(msg, "date_start") {
		t.Fatalf("expected date_start hint, got %q", msg)
	}
}

func TestBuildAccountSummary_compareTotals(t *testing.T) {
	t.Parallel()

	rows := []PacerRow{
		{CurrencyCode: "USD", SpendToDate: 100, SpendPrior: floatPtr(200)},
		{CurrencyCode: "USD", SpendToDate: 50, SpendPrior: floatPtr(100)},
	}
	compare := &pacerCompareContext{periodDays: 3, compareDays: 7}
	summary := buildAccountSummary(rows, defaultPacingThresholds(), compare)
	if summary == nil || summary.SpendPrior == nil || *summary.SpendPrior != 300 {
		t.Fatalf("summary = %+v", summary)
	}
	if summary.SpendChangePercent == nil || *summary.SpendChangePercent != -50 {
		t.Fatalf("change = %v", summary.SpendChangePercent)
	}
	if summary.SpendChangePercentDailyAvg == nil {
		t.Fatal("expected daily avg change on summary")
	}
}
