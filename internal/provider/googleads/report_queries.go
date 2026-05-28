package googleads

import "strings"

func buildKeywordsQuery(in reportFilters) string {
	selectFields := "ad_group_criterion.keyword.text, ad_group_criterion.status, ad_group.id, ad_group.name, campaign.id, campaign.name, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions, metrics.ctr, metrics.average_cpc"
	if in.dateRangeStart != "" || in.dateRangeEnd != "" {
		selectFields += ", segments.date"
	}
	return strings.Join([]string{
		"SELECT " + selectFields,
		"FROM keyword_view",
		googleBuildWhereClause([]string{
			googleInClause("campaign.id", in.campaignIDs),
			googleInClause("ad_group.id", in.adGroupIDs),
			googleEnumInClause("ad_group_criterion.status", in.statuses),
			googleLikeClause("ad_group_criterion.keyword.text", in.keywordContains),
			googleDateBetweenClause(in.dateRangeStart, in.dateRangeEnd),
		}),
		"ORDER BY metrics.cost_micros DESC",
		googleLimitClause(in.limit),
	}, " ")
}

func buildSearchTermsQuery(in reportFilters) string {
	selectFields := "search_term_view.search_term, search_term_view.status, campaign.id, campaign.name, ad_group.id, ad_group.name, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions, metrics.ctr"
	if in.dateRangeStart != "" || in.dateRangeEnd != "" {
		selectFields += ", segments.date"
	}
	return strings.Join([]string{
		"SELECT " + selectFields,
		"FROM search_term_view",
		googleBuildWhereClause([]string{
			googleInClause("campaign.id", in.campaignIDs),
			googleInClause("ad_group.id", in.adGroupIDs),
			googleEnumInClause("campaign.status", in.statuses),
			googleLikeClause("search_term_view.search_term", in.searchTermContains),
			googleDateBetweenClause(in.dateRangeStart, in.dateRangeEnd),
		}),
		"ORDER BY metrics.cost_micros DESC",
		googleLimitClause(in.limit),
	}, " ")
}

func buildPmaxSearchTermsQuery(in reportFilters) string {
	selectFields := "campaign_search_term_view.search_term, campaign.id, campaign.name, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions, metrics.ctr"
	if in.dateRangeStart != "" || in.dateRangeEnd != "" {
		selectFields += ", segments.date"
	}
	return strings.Join([]string{
		"SELECT " + selectFields,
		"FROM campaign_search_term_view",
		googleBuildWhereClause([]string{
			googleInClause("campaign.id", in.campaignIDs),
			googleEnumInClause("campaign.status", in.statuses),
			googleLikeClause("campaign_search_term_view.search_term", in.searchTermContains),
			googleDateBetweenClause(in.dateRangeStart, in.dateRangeEnd),
		}),
		"ORDER BY metrics.cost_micros DESC",
		googleLimitClause(in.limit),
	}, " ")
}

func buildConversionActionsQuery(in reportFilters) string {
	return strings.Join([]string{
		"SELECT conversion_action.id, conversion_action.name, conversion_action.status, conversion_action.type, conversion_action.category, conversion_action.primary_for_goal, conversion_action.include_in_conversions_metric, conversion_action.counting_type",
		"FROM conversion_action",
		googleBuildWhereClause([]string{
			googleEnumInClause("conversion_action.status", in.statuses),
			googleLikeClause("conversion_action.name", in.nameContains),
		}),
		"ORDER BY conversion_action.name ASC",
		googleLimitClause(in.limit),
	}, " ")
}

func buildConversionPerformanceQuery(in reportFilters) string {
	return strings.Join([]string{
		"SELECT campaign.id, campaign.name, segments.conversion_action, metrics.conversions, metrics.conversions_value, metrics.all_conversions, segments.date",
		"FROM campaign",
		googleBuildWhereClause([]string{
			googleInClause("campaign.id", in.campaignIDs),
			googleConversionActionResourceInClause(in.customerID, in.conversionActionIDs),
			googleDateBetweenClause(in.dateRangeStart, in.dateRangeEnd),
		}),
		"ORDER BY metrics.conversions DESC",
		googleLimitClause(in.limit),
	}, " ")
}

func buildOfflineConversionUploadSummariesQuery(in reportFilters) string {
	return strings.Join([]string{
		"SELECT offline_conversion_upload_conversion_action_summary.conversion_action_id, offline_conversion_upload_conversion_action_summary.conversion_action_name, offline_conversion_upload_conversion_action_summary.client, offline_conversion_upload_conversion_action_summary.status, offline_conversion_upload_conversion_action_summary.total_event_count, offline_conversion_upload_conversion_action_summary.successful_event_count, offline_conversion_upload_conversion_action_summary.pending_event_count, offline_conversion_upload_conversion_action_summary.last_upload_date_time",
		"FROM offline_conversion_upload_conversion_action_summary",
		googleBuildWhereClause([]string{
			googleLikeClause("offline_conversion_upload_conversion_action_summary.conversion_action_name", in.nameContains),
		}),
		"ORDER BY offline_conversion_upload_conversion_action_summary.last_upload_date_time DESC",
		googleLimitClause(in.limit),
	}, " ")
}
