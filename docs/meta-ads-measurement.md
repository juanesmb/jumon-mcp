# Meta Ads measurement tools

Agent ladder for pixels, conversions, and signal health. All tools use existing `ads_read` scope (Standard Access) — no new OAuth permissions.

## Workflow

1. **Discover pixels** — `meta_list_datasets` with `act_id` → note `id`, `last_fired_time`, `is_unavailable`.
2. **Inspect one pixel** — `meta_get_dataset` with `dataset_id` for ownership and firing status.
3. **Custom conversions** — `meta_list_custom_conversions` (optional `dataset_id` filter) for optimization events.
4. **Match quality** — `meta_get_dataset_quality` with `dataset_id` for EMQ and coverage (`fields` default: `web{event_match_quality,event_name}`).

## When to use which tool

| Question | Tool |
|----------|------|
| Which pixel is on this account? | `meta_list_datasets` |
| Is this pixel still firing? | `meta_get_dataset` (`last_fired_time`) |
| What custom conversion events exist? | `meta_list_custom_conversions` |
| Why is EMQ low? | `meta_get_dataset_quality` |

## Paused: `meta_get_dataset_stats`

`GET /{dataset_id}/stats` requires `ads_read` **Advanced Access** and returns Permission Denied on Standard Access tokens. Re-enable after App Review. Until then use `last_fired_time` on `meta_get_dataset` and `meta_get_dataset_quality` for signal health.

## Gateway smoke (dataset quality)

```bash
curl -s -X POST -H "x-gateway-secret: $GATEWAY_INTERNAL_SECRET" \
  -H "Content-Type: application/json" \
  "http://localhost:3000/api/internal/providers/meta/proxy" \
  -d '{
    "userId": "USER_CLERK_ID",
    "mcpTool": "meta_get_dataset_quality",
    "method": "GET",
    "path": "dataset_quality",
    "query": {
      "dataset_id": "PIXEL_ID",
      "fields": "web{event_match_quality,event_name}"
    }
  }' | jq .
```

## Related

- [meta-ads-tools.md](meta-ads-tools.md)
- mcp-ads-manager [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md)
