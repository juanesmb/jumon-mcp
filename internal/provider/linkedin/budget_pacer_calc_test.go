package linkedin

import (
	"testing"
	"time"
)

func TestComputePacing_dailyBudgetMidPeriod(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	result := computePacing(PacingInputs{
		BudgetType:   budgetTypeDaily,
		BudgetAmount: 100,
		SpendToDate:  1500,
		PeriodStart:  start,
		PeriodEnd:    end,
	}, defaultPacingThresholds())

	if result.ExpectedSpendToDate == nil || *result.ExpectedSpendToDate != 1000 {
		t.Fatalf("expected = %v", result.ExpectedSpendToDate)
	}
	if result.PacingPercent == nil || *result.PacingPercent != 150 {
		t.Fatalf("pacing = %v", result.PacingPercent)
	}
	if result.PacingStatus != pacingStatusOver {
		t.Fatalf("status = %q", result.PacingStatus)
	}
}

func TestComputePacing_lifetimeProration(t *testing.T) {
	t.Parallel()

	flightStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	flightEnd := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).UnixMilli()
	periodEnd := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	result := computePacing(PacingInputs{
		BudgetType:       budgetTypeLifetime,
		BudgetAmount:     3000,
		RunScheduleStart: &flightStart,
		RunScheduleEnd:   &flightEnd,
		SpendToDate:      800,
		PeriodStart:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:        periodEnd,
	}, defaultPacingThresholds())

	if result.ExpectedSpendToDate == nil {
		t.Fatal("expected spend required")
	}
	// 15/30 of 3000 = 1500
	if *result.ExpectedSpendToDate != 1500 {
		t.Fatalf("expected = %v", *result.ExpectedSpendToDate)
	}
}

func TestComputePacing_unknownWithoutBudget(t *testing.T) {
	t.Parallel()

	result := computePacing(PacingInputs{
		BudgetType:  budgetTypeNone,
		SpendToDate: 50,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC),
	}, defaultPacingThresholds())

	if result.PacingStatus != pacingStatusUnknown {
		t.Fatalf("status = %q", result.PacingStatus)
	}
}

func TestClassifyPacing_customThresholds(t *testing.T) {
	t.Parallel()

	th := PacingThresholds{Over: 1.05, Under: 0.95}
	if classifyPacing(1.06, th) != pacingStatusOver {
		t.Fatal("expected over")
	}
	if classifyPacing(0.94, th) != pacingStatusUnder {
		t.Fatal("expected under")
	}
	if classifyPacing(1.0, th) != pacingStatusOnTrack {
		t.Fatal("expected on_track")
	}
}
