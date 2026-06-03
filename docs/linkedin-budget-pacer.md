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

Set `compare_date_range_start` / `compare_date_range_end` for prior-period spend and WoW-style deltas.

| Field | Meaning |
|-------|---------|
| `spendPrior` | Total spend in the compare period |
| `spendChangePercent` | Raw period-over-period % change (unfair when periods differ in length) |
| `spendChangePercentDailyAvg` | % change using average daily spend (fair for 3d vs 7d, etc.) |

Account `summary` includes the same compare fields when a single currency applies.

**Parameter names:** use `date_range_start` / `date_range_end` (not `start_date` / `end_date`). Aliases `start_date`, `end_date`, and `compare_start_date` are accepted for convenience. Validation errors list accepted parameter names.

## Limitations

- Campaign list capped at **20 pages** of auto-pagination → `metadata.truncated` + warning; narrow with `campaign_ids` or `campaign_group_ids`
- Only campaigns returned by LinkedIn for the authenticated user on that ad account (default `status_filter: ACTIVE`); client/managed accounts outside OAuth scope will not appear
- Account `summary` is omitted when campaigns use mixed currencies (per-row and group rollups still returned)
- Paused campaigns are included when matched by `status_filter`; pacing may be `unknown` but spend is shown

## Future write tools

Campaign updates (pause, budget patch) will live in separate `ToolActionExecute` tools built on `campaign_snapshot.go` — not part of the pacer report.
