# Meta Ads MCP tools

**Status:** P1 + P1.5 + P2 shipped — **14 read tools** via `execute_platform_tool` for platform `meta`.

Manual regression: mcp-ads-manager [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md). API version: [meta-ads-api-version.md](meta-ads-api-version.md).

## Capability matrix

| User need | Tool | Graph API |
|-----------|------|-----------|
| List ad accounts | `meta_list_ad_accounts` | `GET /me?fields=adaccounts{...}` |
| Account details | `meta_get_ad_account` | `GET /{act_id}` |
| List campaigns | `meta_list_campaigns` | `GET /{act_id}/campaigns` |
| Campaign detail | `meta_get_campaign` | `GET /{campaign_id}` |
| List ad sets | `meta_list_ad_sets` | `GET /{act_id}/adsets` |
| Ad set detail | `meta_get_ad_set` | `GET /{adset_id}` |
| List ads | `meta_list_ads` | `GET /{act_id}/ads` |
| Ad detail | `meta_get_ad` | `GET /{ad_id}` |
| Account insights | `meta_get_ad_account_insights` | `GET /{act_id}/insights` |
| Campaign insights | `meta_get_campaign_insights` | `GET /{campaign_id}/insights` |
| Unified performance report | `meta_search_ad_entities` | `GET /{act_id}/insights` + `level` |
| Field metadata (filter/sort) | `meta_get_field_context` | Embedded catalog |
| Delivery errors | `meta_get_delivery_errors` | `GET /{entity_id}?fields=failed_delivery_checks,issues_info,...` |
| Facebook Pages for ads | `meta_list_account_pages` | `GET /me/accounts` |

## Agent workflow

1. Connect Meta in Jumon dashboard.
2. `meta_list_ad_accounts` → pick `act_id` (`act_` prefix or numeric).
3. **Reporting:** prefer `meta_search_ad_entities` with `date_preset` or `time_range`; use `level: adset` or `level: ad` for lower-level metrics.
4. **Structure:** `meta_list_campaigns` → `meta_list_ad_sets` → `meta_list_ads`; use `meta_get_*` for single-object drill-down.
5. **Delivery:** `meta_get_delivery_errors` when ads are not delivering.
6. **Lead Gen:** `meta_list_account_pages` — check `leadgen_tos_accepted` before Lead Gen campaigns.
7. **Placements:** `breakdowns: ["publisher_platform"]` for Facebook vs Instagram.

Time range precedence: `time_ranges` > `time_range` > `since`/`until` > `date_preset` (default `last_30d`).

### Account list note

Official Meta Ads MCP exposes `is_ads_mcp_enabled` / `is_queryable` on accounts. Standard Graph `adaccounts` fields for `ads_read` tokens do not include those flags — handle per-account Graph permission errors in agent responses.

### Ad set / ad insights

No dedicated `meta_get_ad_set_insights` / `meta_get_ad_insights` tools. Use `meta_search_ad_entities` with `level: adset` or `level: ad`.

## Package layout

| File | Role |
|------|------|
| `tools.go` | Registration |
| `schemas.go` | JSON Schema |
| `schema_docs.go` | Agent-facing constants |
| `inputs.go` | Parsers, `normalizeActID` |
| `graph_query.go` | Graph query encoding |
| `pagination.go` | `auto_paginate` (max 10 pages) |
| `accounts.go` | Account list/detail |
| `structure.go` | Campaigns, ad sets, ads |
| `insights.go` | Insights + `meta_search_ad_entities` |
| `delivery.go` | Delivery error batch fetch |
| `pages.go` | Facebook Pages list |
| `field_context.go` | Embedded field catalog |
| `graph_errors.go` | Graph error message parsing |
| `proxy.go` | Gateway port |

## Related

- mcp-ads-manager [meta-ads-oauth.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-oauth.md)
