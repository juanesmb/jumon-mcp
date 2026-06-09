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

Meta (Facebook, Instagram, and other placements) uses one Marketing API. Connect Meta in the Jumon dashboard first.

1. `meta_list_ad_accounts` → pick `act_id` (accepts `act_123` or numeric `123`).
2. **Reporting:** prefer `meta_search_ad_entities` with `date_preset` or `time_range`. Use `level: adset` or `level: ad` for lower-level performance (no separate insight tools). Call `meta_get_field_context` before `filtering` or `sort`.
3. **Structure:** `meta_list_campaigns` → `meta_list_ad_sets` → `meta_list_ads`; drill down with `meta_get_campaign`, `meta_get_ad_set`, `meta_get_ad`.
4. **Delivery issues:** `meta_get_delivery_errors` with campaign/ad set/ad ids when ads are not delivering (uses Graph Batch for multiple ids).
5. **Lead Gen / Pages:** `meta_list_account_pages` — optional `act_id` for promote_pages; check `leadgen_tos_accepted` before Lead Gen work.
6. **Creatives / media:** `meta_list_creatives`, `meta_get_creative`, `meta_get_ad_images`, `meta_get_ad_videos`, `meta_get_ad_preview`.
7. **Targeting / audiences:** `meta_search_interests` → `meta_get_interest_suggestions` → `meta_estimate_audience_size`; `meta_list_custom_audiences` → `meta_get_custom_audience`. Do not use behaviors/demographics browse tools — paused until `ads_read` Advanced Access.
8. **Measurement:** `meta_list_datasets` → `meta_get_dataset` → `meta_list_custom_conversions`; signal health via `meta_get_dataset_quality` and `last_fired_time` on `meta_get_dataset`.
9. **Audit:** `meta_get_account_activities` (account) or `meta_get_ad_set_activities` (one ad set).
10. **Creative governance:** `meta_list_creatives` → `meta_list_creative_ads`.
11. **Optimization:** `meta_get_opportunity_score` for account-level recommendations (not per-campaign).
12. **Placements:** use Insights `breakdowns: ["publisher_platform"]` to split Facebook vs Instagram.
13. `auto_paginate` defaults true on list and insights tools (max 10 pages; insights limit max 1000 per page).
14. API version **v25.0** — see `docs/meta-ads-tools.md`.
