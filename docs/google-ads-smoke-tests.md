# Google Ads MCP smoke tests

Manual regression checklist after Google provider changes. Run against a connected Google Ads account with MCC access (verified with Understory MCC `6599386627`).

## Prerequisites

- Google OAuth connected in Jumon (`usable: true`, `health: active`)
- jumon-mcp running locally or use production MCP URL
- Clerk-authenticated MCP client (Cursor, etc.)

## Account fixtures (reference)

| Account | customer_id | Role |
|---------|-------------|------|
| Understory MCC | `6599386627` | Manager |
| Lightfield (Search) | `1031237849` | Client with Search keywords/search terms |
| Understory client | `1874904215` | Client with Demand Gen/Video only (empty keyword/search term reports) |

When testing a client under MCC, always pass `login_customer_id: "6599386627"`.

## Tool chain checklist

| # | Prompt / action | Expected tool(s) | Pass criteria |
|---|-----------------|------------------|---------------|
| 1 | List my Google ad accounts | `google_list_ad_accounts` | `accounts[]` with `descriptive_name`, `manager` |
| 2 | Find account named "Lightfield" | `google_resolve_customer` | `matches[]` with `customer_id`, optional `login_customer_id` |
| 3 | List campaigns for Lightfield last 30 days | `google_search_campaigns` | `results[]` with campaign + metrics |
| 4 | Keyword report for Lightfield | `google_search_keywords` | Non-empty `results[]` for Search account |
| 5 | Search terms for Lightfield | `google_search_search_terms` | Non-empty `results[]` |
| 6 | Keywords for Understory video client | `google_search_keywords` | Empty `results[]` + `hint` + `channel_summary` (non-Search channels) |
| 7 | PMax search terms for account with PMax | `google_search_pmax_search_terms` | Rows or empty with PMax-specific `hint` |
| 8 | List conversion actions | `google_list_conversion_actions` | `results[]` with conversion action fields |
| 9 | Conversion performance last 30 days | `google_search_conversion_performance` | Defaults date range if omitted |
| 10 | Metadata for keyword_view | `google_get_resource_metadata` | `selectable`, `filterable`, `sortable` arrays |
| 11 | Valid GAQL on campaign_search_term_view | `google_search_gaql` | `campaign.name` + metrics accepted; `_debug.query` present |
| 12 | Invalid GAQL field on view | `google_search_gaql` | Local validation error before API call |
| 13 | Explore platform google | `explore_platform` | 14 Google tools listed |

## Expected response shapes

| Field | When |
|-------|------|
| `hint` | Empty keyword/search term/PMax reports |
| `channel_summary` | Empty reports after channel sniff |
| `truncated: true` | Account list exceeds enrichment cap |
| `_debug.query` | `google_search_gaql` success |
| `metadata.pages_fetched` | Report tools with `auto_paginate: true` and multiple pages |

## GAQL validation regression

**Should pass (attributed fields):**

```json
{
  "customer_id": "1031237849",
  "login_customer_id": "6599386627",
  "resource": "campaign_search_term_view",
  "fields": ["campaign.name", "campaign_search_term_view.search_term", "metrics.clicks"],
  "limit": 10
}
```

**Should fail locally:**

```json
{
  "resource": "campaign_search_term_view",
  "fields": ["campaign.name", "invalid_field"],
  ...
}
```

## Automated coverage

- Unit tests: `go test ./internal/provider/googleads/...`
- CI: `.github/workflows/test.yml` runs on every PR
- Gateway integration smoke: manual until staging credentials are wired in CI

## Related docs

- [google-ads-tools.md](google-ads-tools.md) — capability matrix
- [google-ads-api-version.md](google-ads-api-version.md) — API version upgrades
