# Google Ads MCP tools

Curated GAQL tools plus a hybrid GAQL escape hatch exposed via `execute_platform_tool` for platform `google`. OAuth scope: `https://www.googleapis.com/auth/adwords` (unchanged).

## Capability matrix

| User need | Tool | GAQL resource | Notes |
|-----------|------|---------------|-------|
| List accessible accounts with names | `google_list_ad_accounts` | `customer` (+ listAccessibleCustomers) | Max 50 accounts enriched; skips disabled accounts |
| Resolve account by name | `google_resolve_customer` | `customer`, `customer_client` | Scans up to 5 MCCs when no direct match |
| MCC client list | `google_list_client_accounts_under_manager` | `customer_client` | Optional `client_name_contains` |
| Campaign structure + metrics | `google_search_campaigns` | `campaign` | |
| Ad group structure + metrics | `google_search_ad_groups` | `ad_group` | |
| Ad structure + metrics | `google_search_ads` | `ad_group_ad` | |
| Keyword-level data | `google_search_keywords` | `keyword_view` | Search campaigns |
| Search terms | `google_search_search_terms` | `search_term_view` | Search campaigns |
| Conversion action catalog | `google_list_conversion_actions` | `conversion_action` | Config, no date segments |
| Conversion performance | `google_search_conversion_performance` | `campaign` + `segments.conversion_action` | Defaults to last 30 days |
| Discover GAQL fields | `google_get_resource_metadata` | GoogleAdsFieldService | No `customer_id`; paginated |
| Long-tail GAQL reports | `google_search_gaql` | Any allowlisted resource | Validated fields; metadata-first |

## Hybrid workflow (P2)

```text
1. google_resolve_customer / google_list_ad_accounts
2. Prefer curated tools (keywords, search terms, campaigns, …)
3. If no curated tool fits:
   a. google_get_resource_metadata(resource)
   b. google_search_gaql(fields, resource, conditions, …)
```

Never guess GAQL fields. References:

- [GAQL grammar](https://developers.google.com/google-ads/api/docs/query/grammar)
- [Fields overview (v22)](https://developers.google.com/google-ads/api/fields/v22/overview)
- [Query Builder](https://developers.google.com/google-ads/api/docs/developer-toolkit/gaa-query-builder)
- [Conversions](https://developers.google.com/google-ads/api/docs/conversions/overview)

## vs official Google Ads MCP

| Official MCP | Jumon |
|--------------|-------|
| `list_accessible_customers` (IDs only) | `google_list_ad_accounts` (enriched) + `google_resolve_customer` |
| Generic `search` (any GAQL) | Curated tools + `google_search_gaql` (allowlisted, validated) |
| `get_resource_metadata` | `google_get_resource_metadata` |
| MCP Resources (metrics/segments HTML) | Static doc links in tool descriptions (no MCP Resources) |

Reference: [google-ads-mcp](https://github.com/googleads/google-ads-mcp), [Google Ads API](https://developers.google.com/google-ads/api).

## GAQL allowlist

`internal/provider/googleads/gaql_resources.txt` (~180 resources) is embedded for validation. Re-sync from [official gaql_resources.txt](https://github.com/googleads/google-ads-mcp/blob/main/ads_mcp/gaql_resources.txt) when Google adds resources.

Common resources: `campaign`, `ad_group`, `keyword_view`, `search_term_view`, `campaign_search_term_view`, `conversion_action`, `asset_group`, `shopping_performance_view`, `change_event`, `offline_conversion_upload_conversion_action_summary`.

## Limits

- Account enrichment: 50 accessible customers max (`truncated: true` when capped); disabled accounts omitted with `skipped_unavailable` count.
- MCC client name search: 5 manager accounts max per resolve call.
- Report row limit: default 500, max 1000 (`limit` param); `change_event` max 10000.
- `google_search_gaql` returns `_debug.query` for agent troubleshooting.

## Package layout

| File | Role |
|------|------|
| `tools.go` | Registration |
| `schemas.go` | JSON Schema |
| `schema_docs.go` | Agent-facing descriptions |
| `accounts.go` | List / resolve / MCC clients |
| `reports.go` | Report service methods |
| `report_queries.go` | GAQL query strings for reporting tools |
| `reports_legacy.go` | Campaigns, ad groups, ads |
| `gaql.go` | Shared query builders |
| `gaql_validate.go` | Allowlist + GAQL input validation |
| `gaql_resources.txt` | Embedded resource allowlist |
| `field_service.go` | `google_get_resource_metadata` |
| `generic_search.go` | `google_search_gaql` |
| `inputs.go` | Parsers and shared types |
| `service.go` | `googleSearch`, `googleAdsFieldSearch` |
| `proxy.go` | Gateway port |

## OAuth

No new scopes required. All tools use existing Google Ads API access via the gateway.
