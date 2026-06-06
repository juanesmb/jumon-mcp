# Meta Ads MCP tool reference

## File map

| Concern | File |
|---------|------|
| Tool registration | `tools.go` |
| JSON schemas | `schemas.go` |
| Agent descriptions | `schema_docs.go` |
| Parsers, act_id | `inputs.go` |
| Graph query encoding | `graph_query.go` |
| Graph Batch API | `graph_batch.go` |
| Auto-pagination | `pagination.go` |
| Accounts | `accounts.go` |
| Campaigns, ad sets, ads | `structure.go` |
| Insights + search entities | `insights.go` |
| Creatives, media, targeting, audiences | `creatives.go`, `media.go`, `targeting.go`, `audiences.go` |
| Measurement (pixels, conversions, EMQ) | `measurement.go` |
| Audit trail (activities) | `activities.go` |
| Opportunity score | `intelligence.go` |
| Field catalog | `field_context.go`, `field_catalog_data.go` |
| Gateway port | `proxy.go` |

## Agent workflow

1. `meta_list_ad_accounts` → `act_id`
2. Prefer **`meta_search_ad_entities`** for performance; **`meta_get_field_context`** before filter/sort
3. Structure: `meta_list_campaigns` → `meta_list_ad_sets` → `meta_list_ads`
4. Creatives: `meta_list_creatives` → `meta_get_creative` / `meta_get_ad_preview`
5. Targeting: `meta_search_interests` → `meta_search_behaviors` / `meta_search_demographics` → `meta_get_interest_suggestions` → `meta_estimate_audience_size`
6. Measurement: `meta_list_datasets` → `meta_get_dataset` → `meta_list_custom_conversions`; `meta_get_dataset_stats` / `meta_get_dataset_quality`
7. Audit: `meta_get_account_activities` or `meta_get_ad_set_activities`
8. Optimization: `meta_get_opportunity_score` (account-level only)

## Reference templates

- List + pagination: `meta_list_campaigns`
- Insights: `meta_get_ad_account_insights`
- Unified report: `meta_search_ad_entities`
- Field metadata: `meta_get_field_context`
- Batch delivery errors: `meta_get_delivery_errors` (multiple entity_ids)
- Creatives list: `meta_list_creatives`
- Audience estimate: `meta_estimate_audience_size`

## Graph paths (relative; version in vault)

| Tool | Path |
|------|------|
| `meta_list_ad_accounts` | `me?fields=adaccounts{...}` |
| `meta_get_ad_account` | `{act_id}` |
| `meta_list_campaigns` | `{act_id}/campaigns` |
| `meta_search_ad_entities` | `{act_id}/insights` + `level` |
| `meta_list_creatives` | `{act_id}/adcreatives` |
| `meta_get_ad_images` | `{act_id}/adimages` |
| `meta_get_opportunity_score` | `{act_id}/recommendations` |
| `meta_list_account_pages` | `me/accounts` or `{act_id}/promote_pages` |
| `meta_list_datasets` | `{act_id}/adspixels` |
| `meta_get_dataset` | `{dataset_id}` |
| `meta_list_custom_conversions` | `{act_id}/customconversions` |
| `meta_get_dataset_stats` | `{dataset_id}/stats` |
| `meta_get_dataset_quality` | `dataset_quality?dataset_id=...` |
| `meta_list_creative_ads` | `{creative_id}/ads` |
| `meta_get_account_activities` | `{act_id}/activities` |
| `meta_get_ad_set_activities` | `{adset_id}/activities` |
| `meta_search_behaviors` | `search?type=adTargetingCategory&class=behaviors` |
| `meta_search_demographics` | `search?type=adTargetingCategory&class={class}` |
| `meta_get_interest_suggestions` | `search?type=adinterestsuggestion&interest_list=...` |

## Docs

- [meta-ads-tools.md](../../docs/meta-ads-tools.md)
- [meta-ads-measurement.md](../../docs/meta-ads-measurement.md)
- [meta-ads-intelligence-spike.md](../../docs/meta-ads-intelligence-spike.md)
- [meta-ads-api-version.md](../../docs/meta-ads-api-version.md)
- mcp-ads-manager [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md)
