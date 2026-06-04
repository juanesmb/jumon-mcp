Jumon MCP facade server.

Workflow:
1. Use `explore_platform` to discover platforms and tool summaries.
2. Use `explore_platform` with `tool_names` to load exact tool schemas.
3. Use `execute_platform_tool` with `tool_name` and `tool_parameters`.

Guidelines:
- Batch schema loading in one `explore_platform` call when possible.
- If a platform is disconnected, surface the provided `connect_url` and ask the user to connect it in Jumon.

## LinkedIn workflow

1. Discover account: `linkedin_list_ad_accounts`.
2. **Budget pacing** (spend vs expected, by campaign and campaign group): `linkedin_get_budget_pacer_report` with `account_id` and `date_range_start` / `date_range_end`. Optional `compare_date_range_*` for prior-period spend deltas.
3. Funnel structure: `linkedin_get_campaign_groups` then `linkedin_get_campaigns` when you need raw campaign objects or filters not covered by the pacer.
4. Custom reporting (demographics, conversions, CRM revenue): `linkedin_get_ad_analytics` — not for simple pacing questions.
5. Use `auto_paginate: true` on list tools for large accounts. If the pacer returns `metadata.truncated`, narrow with `campaign_ids` or `campaign_group_ids`.

## Google Ads workflow

1. Start with `google_list_ad_accounts` or `google_resolve_customer` when the user names an account (do not ask for CID if resolve succeeds).
2. When querying a **client account under an MCC**, set `login_customer_id` to the manager customer id.
3. Pass `customer_id` as digits only (strip `customers/` prefix and hyphens).
4. Use `YYYY-MM-DD` for `date_range_start` / `date_range_end` on metric reports.
5. Structure: campaigns → ad groups → ads. Reporting: keywords (`google_search_keywords`), Search search terms (`google_search_search_terms`), PMax search terms (`google_search_pmax_search_terms`), conversions (`google_list_conversion_actions`, `google_search_conversion_performance`), offline upload health (`google_list_offline_conversion_upload_summaries`).
6. Empty keyword/search term results include `hint` and `channel_summary` — often means the account has no Search campaigns; try another client or PMax tool.
7. **Prefer curated tools** for common reports. When no curated tool fits: `google_get_resource_metadata` → `google_search_gaql`. Never guess GAQL fields. Use `auto_paginate: true` (default) for large reports.

## Meta Ads workflow

Meta (Facebook Ads) uses one Marketing API for Facebook, Instagram, and other Meta placements. Connect Meta in the Jumon dashboard first.

- **P0:** MCP read tools are not registered yet — `explore_platform` will not list `meta` until P1. Use the dashboard to connect; gateway smoke uses internal proxy only.
- **Reporting across placements:** use Insights breakdowns such as `publisher_platform` in P1 tools to split Facebook vs Instagram.
- Ad account ids use the `act_` prefix (e.g. `act_1234567890`).
- API version is **v25.0** (see `docs/meta-ads-api-version.md`). Capability matrix: `docs/meta-ads-tools.md`.
