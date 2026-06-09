# Meta Ads intelligence spike

Spike for official Meta Ads MCP intelligence tools vs Graph API availability for Jumon read-only tools.

## Matrix

| Official MCP tool | Graph endpoint | Jumon tool | Status |
|-------------------|----------------|------------|--------|
| `ads_get_opportunity_score` | `GET /{act_id}/recommendations` | `meta_get_opportunity_score` | **Shipped** |
| `ads_get_opportunity_score` (alt) | `GET /{act_id}?fields=opportunity_score` | — | Not needed; recommendations endpoint is richer |
| `ads_insights_performance_trend` | No stable public Graph equivalent for agents | — | **Deferred** (MCP-only / agent-side analysis) |
| `ads_insights_benchmark` | No stable public Graph equivalent | — | **Deferred** |
| `ads_insights_anomaly_detection` | No stable public Graph equivalent | — | **Deferred** |
| `ads_get_business_context` | No Graph read endpoint for ads_read tokens | — | **Deferred** |
| `ads_list_datasets` / pixel read | `GET /{act_id}/adspixels` | `meta_list_datasets`, `meta_get_dataset` | **Shipped** (R1) |
| `ads_get_dataset_stats` | `GET /{dataset_id}/stats` | — | **Paused** — requires `ads_read` Advanced Access |
| `ads_get_dataset_quality` | `GET /dataset_quality` | `meta_get_dataset_quality` | **Shipped** (R2) |
| Custom conversions list | `GET /{act_id}/customconversions` | `meta_list_custom_conversions` | **Shipped** (R1) |
| `adTargetingCategory` browse | `GET /search?type=adTargetingCategory` | — | **Paused** — `meta_search_behaviors`, `meta_search_demographics` need Advanced Access |

## Gateway smoke (opportunity score)

```bash
curl -s -X POST -H "x-gateway-secret: $GATEWAY_INTERNAL_SECRET" \
  -H "Content-Type: application/json" \
  "http://localhost:3000/api/internal/providers/meta/proxy" \
  -d '{
    "userId": "USER_CLERK_ID",
    "mcpTool": "meta_get_opportunity_score",
    "method": "GET",
    "path": "act_ACT_ID/recommendations"
  }' | jq .
```

Expect HTTP 200 with recommendation objects (account-level score and lift values).

## Agent rules

- Opportunity score is **account-level only** — never attribute it to a single campaign or ad.
- Refer to `opportunity_score_lift` as **points**, not "impact" or "percent improvement."
- Call proactively when users ask how to improve performance or what to do next.
- Trend/benchmark/anomaly tools are out of scope until Meta exposes stable Graph read APIs.

## Related

- [meta-ads-tools.md](meta-ads-tools.md)
- mcp-ads-manager [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md)
