# Meta Ads MCP tools

**Status:** P1–P4 shipped — **26 read tools** via `execute_platform_tool` for platform `meta`.

Manual regression: mcp-ads-manager [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md). API version: [meta-ads-api-version.md](meta-ads-api-version.md). Intelligence spike: [meta-ads-intelligence-spike.md](meta-ads-intelligence-spike.md).

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
| Delivery errors | `meta_get_delivery_errors` | Graph Batch GET or single GET per entity |
| Facebook Pages for ads | `meta_list_account_pages` | `GET /me/accounts` or `GET /{act_id}/promote_pages` |
| List creatives | `meta_list_creatives` | `GET /{act_id}/adcreatives` |
| Creative detail | `meta_get_creative` | `GET /{creative_id}` |
| Ad images | `meta_get_ad_images` | `GET /{act_id}/adimages` |
| Ad videos | `meta_get_ad_videos` | `GET /{act_id}/advideos` |
| Ad preview | `meta_get_ad_preview` | `GET /{ad_id}/previews` |
| Search interests | `meta_search_interests` | `GET /search?type=adinterest` |
| Search geo locations | `meta_search_geo_locations` | `GET /search?type=adgeolocation` |
| Estimate audience size | `meta_estimate_audience_size` | `GET /{act_id}/delivery_estimate` |
| List custom audiences | `meta_list_custom_audiences` | `GET /{act_id}/customaudiences` |
| Custom audience detail | `meta_get_custom_audience` | `GET /{custom_audience_id}` |
| Ad sets using audience | `meta_list_custom_audience_ad_sets` | `GET /{custom_audience_id}/adsets` |
| Opportunity score | `meta_get_opportunity_score` | `GET /{act_id}/recommendations` |

## Agent workflow

1. Connect Meta in Jumon dashboard.
2. `meta_list_ad_accounts` → pick `act_id` (`act_` prefix or numeric).
3. **Reporting:** prefer `meta_search_ad_entities` with `date_preset` or `time_range`; use `level: adset` or `level: ad` for lower-level metrics.
4. **Structure:** `meta_list_campaigns` → `meta_list_ad_sets` → `meta_list_ads`; use `meta_get_*` for single-object drill-down.
5. **Creatives / media:** `meta_list_creatives`, `meta_get_creative`, `meta_get_ad_images`, `meta_get_ad_videos`, `meta_get_ad_preview`.
6. **Targeting:** `meta_search_interests` → `meta_search_geo_locations` → `meta_estimate_audience_size`.
7. **Audiences:** `meta_list_custom_audiences` → `meta_get_custom_audience`; before deletion, `meta_list_custom_audience_ad_sets`.
8. **Delivery:** `meta_get_delivery_errors` when ads are not delivering (Graph Batch when multiple ids).
9. **Lead Gen:** `meta_list_account_pages` — optional `act_id` for promote_pages; check `leadgen_tos_accepted`.
10. **Optimization:** `meta_get_opportunity_score` for account-level recommendations (not per-campaign).
11. **Placements:** `breakdowns: ["publisher_platform"]` for Facebook vs Instagram.

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
| `graph_batch.go` | Graph Batch API |
| `pagination.go` | `auto_paginate` (max 10 pages) |
| `accounts.go` | Account list/detail |
| `structure.go` | Campaigns, ad sets, ads |
| `insights.go` | Insights + `meta_search_ad_entities` |
| `delivery.go` | Delivery error batch fetch |
| `pages.go` | Facebook Pages list |
| `field_context.go`, `field_catalog_data.go` | Embedded field catalog |
| `creatives.go`, `media.go`, `targeting.go`, `audiences.go`, `intelligence.go` | P3/P4 tools |
| `graph_errors.go` | Graph error message parsing |
| `proxy.go` | Gateway port (GET + POST form) |

## Related

- mcp-ads-manager [meta-ads-oauth.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-oauth.md)
