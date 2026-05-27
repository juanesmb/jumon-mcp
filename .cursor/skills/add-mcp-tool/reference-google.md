# Google Ads MCP tool reference

## File map

| Concern | File |
|---------|------|
| Tool registration | `tools.go` |
| JSON schemas | `schemas.go` |
| Agent descriptions | `schema_docs.go` |
| Account list / resolve | `accounts.go` |
| Keywords, search terms, conversions | `reports.go`, `report_queries.go` |
| Campaigns, ad groups, ads | `reports_legacy.go` |
| GAQL validation + allowlist | `gaql_validate.go`, `gaql_resources.txt` |
| Field metadata (P2) | `field_service.go` |
| Generic GAQL search (P2) | `generic_search.go` |
| GAQL helpers | `gaql.go` |
| Input parsers | `inputs.go` |
| Search transport | `service.go`, `proxy.go` |

## Hybrid workflow

1. **Prefer curated tools** for common user asks (keywords, search terms, campaigns, conversions).
2. **Escape hatch:** `google_get_resource_metadata` → `google_search_gaql` when no curated tool fits.
3. Never add a new curated tool for one-off GAQL unless it becomes a repeated user pattern.

## GAQL resources

| Tool | FROM / API |
|------|------------|
| `google_search_keywords` | `keyword_view` |
| `google_search_search_terms` | `search_term_view` |
| `google_list_conversion_actions` | `conversion_action` |
| `google_search_conversion_performance` | `campaign` + `segments.conversion_action` |
| `google_get_resource_metadata` | `googleAdsFields:search` |
| `google_search_gaql` | Any resource in `gaql_resources.txt` |

## Conversion reporting

- **Catalog:** `conversion_action` resource — no date segments.
- **Performance:** `campaign` + `segments.conversion_action` + conversion metrics; default last 30 days.

See [docs/google-ads-tools.md](../../docs/google-ads-tools.md) and [Google conversion docs](https://developers.google.com/google-ads/api/docs/conversions/overview).

## Reference templates

- Curated report: `google_search_keywords` or `google_resolve_customer`
- P2 generic: `google_search_gaql` + `field_service.go`
