package meta

const (
	docActID = "Meta ad account id. Accepts act_123456789 or numeric 123456789."

	docAutoPaginate = "When true (default), follows Graph paging.cursors.after for up to 10 pages and merges data[]."

	docInsightsTimePrecedence = "Time range precedence: time_ranges > time_range > since/until > date_preset (default last_30d). Metrics require a time range."

	docPublisherPlatform = "Use breakdowns [\"publisher_platform\"] to split Facebook vs Instagram in Insights responses."

	docNoStandaloneActions = "Do not request standalone actions or action_values fields; use colon action fields such as actions:link_click."

	docSearchEntitiesPreferred = "Prefer meta_search_ad_entities for performance questions at campaign/adset/ad level; use structure list tools for navigation."
)
