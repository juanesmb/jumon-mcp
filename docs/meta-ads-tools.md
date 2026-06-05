# Meta Ads MCP tools

**Status:** P1 + P1.5 shipped — 10 read tools via `execute_platform_tool` for platform `meta`.

Manual regression: mcp-ads-manager [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md). API version: [meta-ads-api-version.md](meta-ads-api-version.md).

## Capability matrix

| User need | Tool | Graph API |
|-----------|------|-----------|
| List ad accounts | `meta_list_ad_accounts` | `GET /me?fields=adaccounts{...}` |
| Account details | `meta_get_ad_account` | `GET /{act_id}` |
| List campaigns | `meta_list_campaigns` | `GET /{act_id}/campaigns` |
| Campaign detail | `meta_get_campaign` | `GET /{campaign_id}` |
| List ad sets | `meta_list_ad_sets` | `GET /{act_id}/adsets` |
| List ads | `meta_list_ads` | `GET /{act_id}/ads` |
| Account insights | `meta_get_ad_account_insights` | `GET /{act_id}/insights` |
| Campaign insights | `meta_get_campaign_insights` | `GET /{campaign_id}/insights` |
| Unified performance report | `meta_search_ad_entities` | `GET /{act_id}/insights` + `level` |
| Field metadata (filter/sort) | `meta_get_field_context` | Embedded catalog |

## Agent workflow

1. Connect Meta in Jumon dashboard.
2. `meta_list_ad_accounts` → pick `act_id` (`act_` prefix or numeric).
3. **Reporting:** prefer `meta_search_ad_entities` with `date_preset` or `time_range`; call `meta_get_field_context` before `filtering` / `sort`.
4. **Structure:** `meta_list_campaigns` → `meta_list_ad_sets` → `meta_list_ads` for navigation.
5. **Placements:** `breakdowns: ["publisher_platform"]` for Facebook vs Instagram.

Time range precedence: `time_ranges` > `time_range` > `since`/`until` > `date_preset` (default `last_30d`).

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
| `field_context.go` | Embedded field catalog |
| `proxy.go` | Gateway port |

## Related

- mcp-ads-manager [meta-ads-oauth.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-oauth.md)
