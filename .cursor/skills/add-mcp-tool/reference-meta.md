# Meta Ads MCP tool reference

## File map

| Concern | File |
|---------|------|
| Tool registration | `tools.go` |
| JSON schemas | `schemas.go` |
| Agent descriptions | `schema_docs.go` |
| Parsers, act_id | `inputs.go` |
| Graph query encoding | `graph_query.go` |
| Auto-pagination | `pagination.go` |
| Accounts | `accounts.go` |
| Campaigns, ad sets, ads | `structure.go` |
| Insights + search entities | `insights.go` |
| Field catalog | `field_context.go` |
| Gateway port | `proxy.go` |

## Agent workflow

1. `meta_list_ad_accounts` → `act_id`
2. Prefer **`meta_search_ad_entities`** for performance; **`meta_get_field_context`** before filter/sort
3. Structure: `meta_list_campaigns` → `meta_list_ad_sets` → `meta_list_ads`

## Reference templates

- List + pagination: `meta_list_campaigns`
- Insights: `meta_get_ad_account_insights`
- Unified report: `meta_search_ad_entities`
- Field metadata: `meta_get_field_context`

## Graph paths (relative; version in vault)

| Tool | Path |
|------|------|
| `meta_list_ad_accounts` | `me?fields=adaccounts{...}` |
| `meta_get_ad_account` | `{act_id}` |
| `meta_list_campaigns` | `{act_id}/campaigns` |
| `meta_search_ad_entities` | `{act_id}/insights` + `level` |

## Docs

- [meta-ads-tools.md](../../docs/meta-ads-tools.md)
- [meta-ads-api-version.md](../../docs/meta-ads-api-version.md)
- mcp-ads-manager [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md)
