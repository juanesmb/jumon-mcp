package meta

const (
	docActID = "Meta ad account id. Accepts act_123456789 or numeric 123456789."

	docAutoPaginate = "When true (default), follows Graph paging.cursors.after for up to 10 pages and merges data[]. When false, metadata.has_more is set when paging.next exists."

	docInsightsTimePrecedence = "Time range precedence: time_ranges > time_range > since/until > date_preset (default last_30d). Metrics require a time range."

	docPublisherPlatform = "Use breakdowns [\"publisher_platform\"] to split Facebook vs Instagram in Insights responses."

	docNoStandaloneActions = "Do not request standalone actions or action_values fields. Use the clicks metric for link-click counts; colon action fields are not valid in default insights fields."

	docSearchEntitiesPreferred = "Prefer meta_search_ad_entities for performance questions at campaign/adset/ad level; use structure list tools for navigation."

	docInsightsLevels = "For ad set or ad performance, use meta_search_ad_entities with level adset or ad (not separate insight tools)."

	docAccountListNote = "Lists accounts via GET /me?fields=adaccounts{...}. Official Meta MCP is_ads_mcp_enabled/is_queryable flags are not exposed on standard Graph adaccounts fields."

	docLeadGenTOS = "When leadgen_tos_accepted is false, direct the user to https://www.facebook.com/ads/leadgen/tos before Lead Gen campaigns."

	docAccountPagesScope = "Without act_id: GET /me/accounts (user/token-scoped Pages). With act_id: GET /{act_id}/promote_pages (Pages promotable from that ad account). Reconnect with pages_show_list if Page fields are missing."

	docDatasetID = "Meta pixel/dataset id (numeric). Same value returned by meta_list_datasets."

	docMeasurementWorkflow = "Measurement ladder: meta_list_datasets → meta_get_dataset → meta_list_custom_conversions; signal health via meta_get_dataset_stats / meta_get_dataset_quality."

	docActivitiesTime = "time_range overrides since/until. Default Graph window is ~1 week when no dates are set."

	docInterestSuggestions = "Pass interest_list as seed names (e.g. [\"Basketball\"]); returns related interests for targeting research."

	docDemographicClass = "Targeting category class. Default demographics. Also: life_events, industries, income, family_statuses, user_device, user_os."

	docDatasetQualityNote = "Dataset Quality API returns EMQ and coverage for a pixel. Requires ads_read on the connected user token."
)
