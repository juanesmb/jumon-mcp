package linkedin

import (
	"math"
	"time"
)

const (
	pacingStatusOnTrack = "on_track"
	pacingStatusOver    = "over"
	pacingStatusUnder   = "under"
	pacingStatusUnknown = "unknown"
)

// PacingThresholds configures over/under classification.
type PacingThresholds struct {
	Over  float64
	Under float64
}

func defaultPacingThresholds() PacingThresholds {
	return PacingThresholds{Over: 1.1, Under: 0.9}
}

func normalizePacingThresholds(in PacingThresholds) PacingThresholds {
	out := defaultPacingThresholds()
	if in.Over > 0 {
		out.Over = in.Over
	}
	if in.Under > 0 {
		out.Under = in.Under
	}
	return out
}

// PacingInputs holds one campaign's budget config and observed spend for a period.
type PacingInputs struct {
	BudgetType       string
	BudgetAmount     float64
	RunScheduleStart *int64
	RunScheduleEnd   *int64
	SpendToDate      float64
	PeriodStart      time.Time
	PeriodEnd        time.Time
}

// PacingResult is computed pacing metrics for one campaign or rollup.
type PacingResult struct {
	ExpectedSpendToDate  *float64
	PacingPercent        *float64
	PacingStatus         string
	ProjectedPeriodSpend *float64
}

func computePacing(in PacingInputs, thresholds PacingThresholds) PacingResult {
	thresholds = normalizePacingThresholds(thresholds)
	expected, ok := expectedSpendToDate(in)
	if !ok || expected <= 0 {
		return PacingResult{PacingStatus: pacingStatusUnknown}
	}

	result := PacingResult{
		ExpectedSpendToDate: floatPtr(expected),
		PacingStatus:        pacingStatusUnknown,
	}

	if in.SpendToDate >= 0 {
		pct := (in.SpendToDate / expected) * 100
		result.PacingPercent = floatPtr(round2(pct))
		result.PacingStatus = classifyPacing(pct/100, thresholds)
	}

	periodDays := inclusiveUTCDays(in.PeriodStart, in.PeriodEnd)
	elapsed := inclusiveUTCDays(in.PeriodStart, in.PeriodEnd)
	if periodDays > 0 && elapsed > 0 && in.SpendToDate >= 0 {
		projected := (in.SpendToDate / float64(elapsed)) * float64(periodDays)
		result.ProjectedPeriodSpend = floatPtr(round2(projected))
	}

	return result
}

func expectedSpendToDate(in PacingInputs) (float64, bool) {
	elapsed := inclusiveUTCDays(in.PeriodStart, in.PeriodEnd)
	if elapsed <= 0 || in.BudgetAmount <= 0 {
		return 0, false
	}

	switch in.BudgetType {
	case budgetTypeDaily:
		return in.BudgetAmount * float64(elapsed), true
	case budgetTypeLifetime:
		if in.RunScheduleEnd == nil {
			return 0, false
		}
		flightStart := in.PeriodStart
		if in.RunScheduleStart != nil {
			flightStart = time.UnixMilli(*in.RunScheduleStart).UTC()
			flightStart = time.Date(flightStart.Year(), flightStart.Month(), flightStart.Day(), 0, 0, 0, 0, time.UTC)
		}
		flightEnd := time.UnixMilli(*in.RunScheduleEnd).UTC()
		flightEnd = time.Date(flightEnd.Year(), flightEnd.Month(), flightEnd.Day(), 0, 0, 0, 0, time.UTC)
		totalFlightDays := inclusiveUTCDays(flightStart, flightEnd)
		if totalFlightDays <= 0 {
			return 0, false
		}
		return in.BudgetAmount * (float64(elapsed) / float64(totalFlightDays)), true
	default:
		return 0, false
	}
}

func classifyPacing(ratio float64, thresholds PacingThresholds) string {
	if ratio >= thresholds.Over {
		return pacingStatusOver
	}
	if ratio <= thresholds.Under {
		return pacingStatusUnder
	}
	return pacingStatusOnTrack
}

func inclusiveUTCDays(start, end time.Time) int {
	start = truncateUTCDate(start)
	end = truncateUTCDate(end)
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func truncateUTCDate(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func floatPtr(v float64) *float64 {
	return &v
}

func spendChangePercent(current, prior float64) *float64 {
	if prior <= 0 {
		return nil
	}
	pct := ((current - prior) / prior) * 100
	return floatPtr(round2(pct))
}

// spendChangePercentDailyAvg compares average daily spend across periods of different lengths.
func spendChangePercentDailyAvg(current, prior float64, currentDays, priorDays int) *float64 {
	if prior <= 0 || currentDays <= 0 || priorDays <= 0 {
		return nil
	}
	priorDaily := prior / float64(priorDays)
	if priorDaily <= 0 {
		return nil
	}
	currentDaily := current / float64(currentDays)
	pct := ((currentDaily - priorDaily) / priorDaily) * 100
	return floatPtr(round2(pct))
}

func pacerSpendCompareFields(spend, prior float64, compare *pacerCompareContext) (spendPrior, changePct, changeDaily *float64) {
	p := round2(prior)
	spendPrior = &p
	changePct = spendChangePercent(spend, prior)
	if compare != nil && compare.compareDays > 0 {
		changeDaily = spendChangePercentDailyAvg(spend, prior, compare.periodDays, compare.compareDays)
	}
	return spendPrior, changePct, changeDaily
}
