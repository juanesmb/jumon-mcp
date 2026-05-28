package googleads

import "strings"

func googleAdsWorkflowHint() string {
	return `Google Ads workflow: (1) google_list_ad_accounts or google_resolve_customer when the user names an account; (2) set login_customer_id when querying a client under an MCC; (3) pass customer_id as digits only (no hyphens or customers/ prefix); (4) use YYYY-MM-DD date ranges for metric reports.`
}

func googleAdsHybridHint() string {
	return ` Prefer curated google_* tools for common reports. When no curated tool fits: google_get_resource_metadata then google_search_gaql. Never guess GAQL fields. References: GAQL grammar https://developers.google.com/google-ads/api/docs/query/grammar — fields https://developers.google.com/google-ads/api/fields/v22/overview — Query Builder https://developers.google.com/google-ads/api/docs/developer-toolkit/gaa-query-builder — conversions https://developers.google.com/google-ads/api/docs/conversions/overview`
}

func listAdAccountsDescription() string {
	return googleAdsWorkflowHint() + " Returns accessible accounts with descriptive_name, manager flag, currency, and timezone."
}

func resolveCustomerDescription() string {
	return googleAdsWorkflowHint() + " Matches account_name against direct accessible accounts and optionally client accounts under up to 10 manager (MCC) accounts."
}

func listClientAccountsDescription() string {
	return googleAdsWorkflowHint() + " Lists non-manager client accounts under one manager account. customer_id should be the MCC id; login_customer_id defaults to customer_id when omitted."
}

func searchKeywordsDescription() string {
	return googleAdsWorkflowHint() + " Keyword-level performance from keyword_view. Requires Search campaigns with keywords. Use keyword_contains, campaign_ids, ad_group_ids, statuses, and date_range_start/end filters. When results are empty, the response includes hint and channel_summary — common cause is a non-Search account (Demand Gen, Video, PMax only); try google_resolve_customer for a Search client or google_search_pmax_search_terms for PMax."
}

func searchSearchTermsDescription() string {
	return googleAdsWorkflowHint() + searchTermsToolSelectionHint() + " Search query report from search_term_view. Requires Search campaigns. Use search_term_contains, campaign_ids, ad_group_ids, and date_range_start/end filters. Empty results include hint and channel_summary; for PMax search terms use google_search_pmax_search_terms instead."
}

func searchPmaxSearchTermsDescription() string {
	return googleAdsWorkflowHint() + " Performance Max search term report from campaign_search_term_view. Use search_term_contains, campaign_ids, and date_range_start/end filters. Empty results include hint and channel_summary; confirm the account has Performance Max campaigns."
}

func listConversionActionsDescription() string {
	return googleAdsWorkflowHint() + " Lists conversion action definitions (names, types, status). Config resource — no date segments."
}

func searchConversionPerformanceDescription() string {
	return googleAdsWorkflowHint() + " Conversion metrics by campaign and conversion action. Defaults to last 30 days when no date range is provided. Optional conversion_action_ids (numeric ids or full resource names). Example: customer_id + login_customer_id for MCC clients; omit date_range to use last 30 days."
}

func listOfflineConversionUploadSummariesDescription() string {
	return googleAdsWorkflowHint() + " Offline conversion upload status per conversion action (event counts, status, last upload time). Optional name_contains filter on conversion_action_name."
}

func searchTermsToolSelectionHint() string {
	return " Search terms: use google_search_search_terms for Search campaigns; google_search_pmax_search_terms for Performance Max (campaign_search_term_view)."
}

func getResourceMetadataDescription() string {
	return googleAdsWorkflowHint() + googleAdsHybridHint() +
		" Returns selectable, filterable, and sortable fields for a GAQL resource including compatible metrics.* and segments.* names. Call before google_search_gaql when fields are unknown."
}

func searchGAQLDescription() string {
	common := strings.Join(commonGAQLResources(), ", ")
	return googleAdsWorkflowHint() + googleAdsHybridHint() +
		" Runs a validated GAQL query when no curated tool fits. Use google_get_resource_metadata first. Attributed fields (campaign.*, ad_group.*, etc.) are allowed on views when compatible; unknown fields fail local validation. Common resources: " + common +
		". Full allowlist in docs/google-ads-tools.md. auto_paginate defaults true (up to 10 pages)."
}
