package googleads

func googleAdsWorkflowHint() string {
	return `Google Ads workflow: (1) google_list_ad_accounts or google_resolve_customer when the user names an account; (2) set login_customer_id when querying a client under an MCC; (3) pass customer_id as digits only (no hyphens or customers/ prefix); (4) use YYYY-MM-DD date ranges for metric reports.`
}

func listAdAccountsDescription() string {
	return googleAdsWorkflowHint() + " Returns accessible accounts with descriptive_name, manager flag, currency, and timezone."
}

func resolveCustomerDescription() string {
	return googleAdsWorkflowHint() + " Matches account_name against direct accessible accounts and optionally client accounts under up to 5 manager (MCC) accounts."
}

func listClientAccountsDescription() string {
	return googleAdsWorkflowHint() + " Lists non-manager client accounts under one manager account. customer_id should be the MCC id; login_customer_id defaults to customer_id when omitted."
}

func searchKeywordsDescription() string {
	return googleAdsWorkflowHint() + " Keyword-level performance from keyword_view. Requires Search campaigns. Use keyword_contains, campaign_ids, ad_group_ids, statuses, and date_range_start/end filters."
}

func searchSearchTermsDescription() string {
	return googleAdsWorkflowHint() + " Search query report from search_term_view. Requires Search campaigns. Use search_term_contains, campaign_ids, ad_group_ids, and date_range_start/end filters."
}

func listConversionActionsDescription() string {
	return googleAdsWorkflowHint() + " Lists conversion action definitions (names, types, status). Config resource — no date segments."
}

func searchConversionPerformanceDescription() string {
	return googleAdsWorkflowHint() + " Conversion metrics by campaign and conversion action. Defaults to last 30 days when no date range is provided. Optional conversion_action_ids (numeric ids or full resource names)."
}
