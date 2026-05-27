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
| GAQL helpers | `gaql.go` |
| Input parsers | `inputs.go` |
| Search transport | `service.go`, `proxy.go` |

## GAQL resources (P1)

| Tool | FROM |
|------|------|
| `google_search_keywords` | `keyword_view` |
| `google_search_search_terms` | `search_term_view` |
| `google_list_conversion_actions` | `conversion_action` |
| `google_search_conversion_performance` | `campaign` (with `segments.conversion_action`) |

## Conversion reporting

- **Catalog:** `conversion_action` resource — no date segments.
- **Performance:** `campaign` + `segments.conversion_action` + conversion metrics; default last 30 days.

See [docs/google-ads-tools.md](../../docs/google-ads-tools.md) and [Google conversion docs](https://developers.google.com/google-ads/api/docs/conversions/overview).

## Reference template

Copy patterns from `google_search_keywords` or `google_resolve_customer` in `internal/provider/googleads/`.
