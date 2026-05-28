Jumon MCP facade server.

Workflow:
1. Use `explore_platform` to discover platforms and tool summaries.
2. Use `explore_platform` with `tool_names` to load exact tool schemas.
3. Use `execute_platform_tool` with `tool_name` and `tool_parameters`.

Guidelines:
- Batch schema loading in one `explore_platform` call when possible.
- If a platform is disconnected, surface the provided `connect_url` and ask the user to connect it in Jumon.

## Google Ads workflow

1. Start with `google_list_ad_accounts` or `google_resolve_customer` when the user names an account (do not ask for CID if resolve succeeds).
2. When querying a **client account under an MCC**, set `login_customer_id` to the manager customer id.
3. Pass `customer_id` as digits only (strip `customers/` prefix and hyphens).
4. Use `YYYY-MM-DD` for `date_range_start` / `date_range_end` on metric reports.
5. Structure: campaigns → ad groups → ads. Reporting: keywords (`google_search_keywords`), Search search terms (`google_search_search_terms`), PMax search terms (`google_search_pmax_search_terms`), conversions (`google_list_conversion_actions`, `google_search_conversion_performance`), offline upload health (`google_list_offline_conversion_upload_summaries`).
6. Empty keyword/search term results include `hint` and `channel_summary` — often means the account has no Search campaigns; try another client or PMax tool.
7. **Prefer curated tools** for common reports. When no curated tool fits: `google_get_resource_metadata` → `google_search_gaql`. Never guess GAQL fields. Use `auto_paginate: true` (default) for large reports.
