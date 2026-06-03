# LinkedIn budget pacer (MCP)

Read-only composite tool: **`linkedin_get_budget_pacer_report`**.

## When to use

| Question | Tool |
|----------|------|
| How are we pacing vs budget this month? | `linkedin_get_budget_pacer_report` |
| Custom metrics, demographics, conversion pivots | `linkedin_get_ad_analytics` |
| Raw campaign budget fields only | `linkedin_get_campaigns` |

## Data sources

LinkedIn does not expose a single “pacing” API. Jumon combines:

1. **Campaign config** — `GET adAccounts/{id}/adCampaigns` (`dailyBudget`, `totalBudget`, `runSchedule`, `status`)
2. **Spend** — `GET adAnalytics` (`q=analytics`, pivot `CAMPAIGN`, `timeGranularity=ALL`, `costInLocalCurrency`)

## Pacing formulas (UTC)

- **Daily budget:** `expectedSpendToDate = dailyAmount × elapsedDaysInPeriod`
- **Lifetime budget** (with `runSchedule.end`): prorate `totalBudget` over flight days
- **No budget / no end date:** `pacingStatus: unknown` (spend still returned)
- **`pacingPercent`:** `spendToDate / expectedSpendToDate × 100` when expected > 0
- **`projectedPeriodSpend`:** linear extrapolation `(spend / elapsedDays) × periodDays` — not LinkedIn delivery pacing

Default thresholds: **over** ≥ 110%, **under** ≤ 90% of expected (configurable via `pacing_thresholds`).

## Optional compare period

Set `compare_date_range_start` / `compare_date_range_end` for `spendPrior` and `spendChangePercent` per row.

## Limitations

- Campaign list capped at **20 pages** of auto-pagination → `metadata.truncated` + warning; narrow with `campaign_ids` or `campaign_group_ids`
- Account `summary` is omitted when campaigns use mixed currencies (per-row and group rollups still returned)
- Paused campaigns are included when matched by `status_filter`; pacing may be `unknown` but spend is shown

## Future write tools

Campaign updates (pause, budget patch) will live in separate `ToolActionExecute` tools built on `campaign_snapshot.go` — not part of the pacer report.
