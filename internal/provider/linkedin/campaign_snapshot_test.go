package linkedin

import "testing"

func TestParseCampaignSnapshot_dailyBudget(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"id":            float64(12345),
		"name":          "MOFU - Retargeting",
		"status":        "ACTIVE",
		"campaignGroup": "urn:li:sponsoredCampaignGroup:999",
		"dailyBudget": map[string]any{
			"amount":       "100.50",
			"currencyCode": "USD",
		},
		"runSchedule": map[string]any{
			"start": float64(1_700_000_000_000),
			"end":   float64(1_800_000_000_000),
		},
	}

	snap, ok := parseCampaignSnapshot(row)
	if !ok {
		t.Fatal("expected ok")
	}
	if snap.ID != "12345" {
		t.Fatalf("id = %q", snap.ID)
	}
	if snap.BudgetType != budgetTypeDaily {
		t.Fatalf("budgetType = %q", snap.BudgetType)
	}
	if snap.BudgetAmount != 100.50 {
		t.Fatalf("budgetAmount = %v", snap.BudgetAmount)
	}
	if snap.CurrencyCode != "USD" {
		t.Fatalf("currency = %q", snap.CurrencyCode)
	}
	if snap.CampaignGroupURN != "urn:li:sponsoredCampaignGroup:999" {
		t.Fatalf("group = %q", snap.CampaignGroupURN)
	}
	if snap.RunScheduleStart == nil || snap.RunScheduleEnd == nil {
		t.Fatal("expected run schedule bounds")
	}
}

func TestParseCampaignSnapshot_lifetimeBudget(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"id":     "99",
		"status": "ACTIVE",
		"totalBudget": map[string]any{
			"amount":       "5000",
			"currencyCode": "EUR",
		},
	}

	snap, ok := parseCampaignSnapshot(row)
	if !ok {
		t.Fatal("expected ok")
	}
	if snap.BudgetType != budgetTypeLifetime {
		t.Fatalf("budgetType = %q", snap.BudgetType)
	}
	if snap.BudgetAmount != 5000 {
		t.Fatalf("amount = %v", snap.BudgetAmount)
	}
}

func TestFilterSnapshotsByCampaignIDs(t *testing.T) {
	t.Parallel()

	snapshots := []CampaignSnapshot{
		{ID: "1", URN: "urn:li:sponsoredCampaign:1"},
		{ID: "2", URN: "urn:li:sponsoredCampaign:2"},
	}
	filtered := filterSnapshotsByCampaignIDs(snapshots, []string{"urn:li:sponsoredCampaign:2"})
	if len(filtered) != 1 || filtered[0].ID != "2" {
		t.Fatalf("filtered = %+v", filtered)
	}
}
