# Google Ads MCP tools

Curated GAQL tools exposed via `execute_platform_tool` for platform `google`. OAuth scope: `https://www.googleapis.com/auth/adwords` (unchanged by P1).

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

## vs official Google Ads MCP

| Official MCP | Jumon P1 |
|--------------|----------|
| `list_accessible_customers` (IDs only) | `google_list_ad_accounts` (enriched) + `google_resolve_customer` |
| Generic `search` (any GAQL) | Curated tools with validated filters |
| `get_resource_metadata` | Not in P1 (P2) |

Reference: [google-ads-mcp](https://github.com/googleads/google-ads-mcp), [Google Ads API](https://developers.google.com/google-ads/api).

## Account discovery workflow

```text
User names account → google_resolve_customer
  → use customer_id + login_customer_id (if MCC client)
  → reporting / structure tools
```

If resolve returns no matches, run `google_list_ad_accounts` and ask the user to pick from `descriptive_name` values returned.

## Limits

- Account enrichment: 50 accessible customers max (`truncated: true` when capped); disabled accounts omitted with `skipped_unavailable` count.
- MCC client name search: 5 manager accounts max per resolve call.
- Report row limit: default 500, max 1000 (`limit` param).

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
| `gaql.go` | Query builders |
| `inputs.go` | Parsers and shared types |
| `service.go` | `googleSearch` helper |
| `proxy.go` | Gateway port |

## OAuth

No new scopes required for P1. All tools use existing Google Ads API access via the gateway.
