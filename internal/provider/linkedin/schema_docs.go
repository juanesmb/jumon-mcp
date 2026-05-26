package linkedin

import "strings"

// LinkedIn adAnalytics field names (finder_type=analytics). Any valid API field may be
// requested; this catalog improves agent discoverability. Max 20 fields per API call.
var (
	linkedInAnalyticsFieldCatalog = []struct {
		category string
		fields   []string
	}{
		{
			category: "Identity & time",
			fields:   []string{"pivotValues", "dateRange"},
		},
		{
			category: "Delivery / brand awareness (date range <= 92 days; non-MEMBER_ pivots)",
			fields: []string{
				"approximateMemberReach", // Campaign Manager: Reach
				"audiencePenetration",
				"impressions",
				// averageFrequency: derived by this connector (impressions / approximateMemberReach)
			},
		},
		{
			category: "Cost & spend",
			fields: []string{
				"costInLocalCurrency",
				"costInUsd",
				"costPerQualifiedLead",
				"conversionValueInLocalCurrency",
			},
		},
		{
			category: "Clicks & site traffic",
			fields: []string{
				"clicks",
				"landingPageClicks",
				"companyPageClicks",
				"textUrlClicks",
				"downloadClicks",
			},
		},
		{
			category: "Engagement (Sponsored Content)",
			fields: []string{
				"totalEngagements",
				"likes",
				"comments",
				"commentLikes",
				"shares",
				"follows",
				"reactions",
				"otherEngagements",
				"subscriptionClicks",
			},
		},
		{
			category: "Video",
			fields: []string{
				"videoViews",
				"videoStarts",
				"videoCompletions",
				"videoFirstQuartileCompletions",
				"videoMidpointCompletions",
				"videoThirdQuartileCompletions",
				"averageVideoWatchTime",
				"fullScreenPlays",
			},
		},
		{
			category: "Lead generation",
			fields: []string{
				"oneClickLeads",
				"oneClickLeadFormOpens",
				"qualifiedLeads",
				"validWorkEmailLeads",
				"talentLeads",
				"leadGenerationMailContactInfoShares",
				"leadGenerationMailInterestedClicks",
			},
		},
		{
			category: "Sponsored messaging (Convo / InMail)",
			fields: []string{
				"sends",
				"opens",
				"actionClicks",
				"adUnitClicks",
				"headlineClicks",
				"headlineImpressions",
			},
		},
		{
			category: "Conversions",
			fields: []string{
				"externalWebsiteConversions",
				"externalWebsitePostClickConversions",
				"externalWebsitePostViewConversions",
			},
		},
		{
			category: "Document ads",
			fields: []string{
				"documentCompletions",
				"documentFirstQuartileCompletions",
				"documentMidpointCompletions",
				"documentThirdQuartileCompletions",
			},
		},
		{
			category: "Carousel",
			fields: []string{
				"cardClicks",
				"cardImpressions",
			},
		},
		{
			category: "Jobs & events",
			fields: []string{
				"jobApplications",
				"jobApplyClicks",
				"registrations",
				"postClickRegistrations",
				"postViewRegistrations",
			},
		},
		{
			category: "Revenue attribution (CRM; finder_type=attributedRevenueMetrics; requires CRM connected to LinkedIn). Request revenueAttributionMetrics (returns all nested metrics) or revenueAttributionMetrics:(revenueWonInUsd,returnOnAdSpend,...). Nested metrics:",
			fields: []string{
				"revenueAttributionMetrics",
				"revenueWonInUsd",
				"returnOnAdSpend",
				"closedWonOpportunities",
				"openOpportunities",
				"opportunityAmountInUsd",
				"opportunityWinRate",
				"averageDealSizeInUsd",
				"averageDaysToClose",
			},
		},
	}

	linkedInAnalyticsFieldExamples = [][]string{
		// Brand awareness / delivery (WoW reports)
		{
			"pivotValues",
			"approximateMemberReach",
			"audiencePenetration",
			"impressions",
			"costInLocalCurrency",
			"videoViews",
			"clicks",
			"totalEngagements",
		},
		// Website visits / MOFU traffic
		{
			"pivotValues",
			"impressions",
			"clicks",
			"landingPageClicks",
			"costInLocalCurrency",
			"externalWebsiteConversions",
		},
		// Conversion performance — call linkedin_list_conversions first to get rule IDs,
		// then pivot by CAMPAIGN or CREATIVE to break down by funnel stage.
		{
			"pivotValues",
			"impressions",
			"clicks",
			"externalWebsiteConversions",
			"externalWebsitePostClickConversions",
			"externalWebsitePostViewConversions",
			"conversionValueInLocalCurrency",
			"costInLocalCurrency",
		},
		// Lead gen / incentive
		{
			"pivotValues",
			"impressions",
			"oneClickLeads",
			"oneClickLeadFormOpens",
			"qualifiedLeads",
			"costInLocalCurrency",
			"totalEngagements",
		},
		// Sponsored messaging / BOFU
		{
			"pivotValues",
			"sends",
			"opens",
			"actionClicks",
			"clicks",
			"costInLocalCurrency",
			"oneClickLeads",
		},
		// CRM revenue attribution (HubSpot, Salesforce, Dynamics via LinkedIn)
		{
			"pivotValues",
			"dateRange",
			"revenueAttributionMetrics",
		},
		// Demographic breakdown (no reach on MEMBER_* pivots)
		{
			"pivotValues",
			"impressions",
			"clicks",
			"costInLocalCurrency",
		},
	}
)

func analyticsFieldsSchemaProperty() map[string]any {
	return map[string]any{
		"type":        "array",
		"items":       map[string]any{"type": "string"},
		"description": buildAnalyticsFieldsDescription(),
		"examples":    linkedInAnalyticsFieldExamples,
	}
}

func buildAnalyticsFieldsDescription() string {
	var b strings.Builder
	b.WriteString("LinkedIn adAnalytics metric field names to return (max 20 per request). ")
	b.WriteString("Pass any valid API field; names below are the supported catalog. ")
	b.WriteString("When pivots are set on finder_type analytics, pivotValues is auto-included; reach and impressions are auto-included for non-MEMBER_* pivots; demographic pivots default to impressions, clicks, and costInLocalCurrency. ")
	b.WriteString("For finder_type attributedRevenueMetrics, defaults to pivotValues, dateRange, and revenueAttributionMetrics; do not pass time_granularity (not supported). ")
	b.WriteString("Revenue metric names (revenueWonInUsd, returnOnAdSpend, etc.) are nested inside revenueAttributionMetrics, not top-level field projections. ")
	b.WriteString("Response may include averageFrequency (derived: impressions / approximateMemberReach, Campaign Manager Average frequency).\n\n")

	for _, cat := range linkedInAnalyticsFieldCatalog {
		b.WriteString(cat.category)
		b.WriteString(": ")
		b.WriteString(strings.Join(cat.fields, ", "))
		b.WriteString(".\n")
	}

	b.WriteString("\nConstraints: reach metrics and averageFrequency need date range <= 92 days and non-MEMBER_* pivots; ")
	b.WriteString("MEMBER_* demographic pivots return top 100 values per creative per day and suppress values with fewer than 3 events. ")
	b.WriteString("MEMBER_INDUSTRY rows include pivotLabels (human-readable names) alongside pivotValues URNs. ")
	b.WriteString("Use time_granularity ALL for weekly totals (do not sum daily reach). ")
	b.WriteString("For CRM revenue metrics use finder_type attributedRevenueMetrics (requires CRM connected to LinkedIn; date range 30–366 days; no time_granularity; pivots ACCOUNT, CAMPAIGN_GROUP, or CAMPAIGN only; openOpportunities and opportunityAmountInUsd only when date_range_end is today UTC). ")
	b.WriteString("Viral variants exist (viralImpressions, viralVideoViews, etc.) — prefix viral to the paid metric name.")

	return b.String()
}

func linkedInGetAdAnalyticsToolDescription() string {
	return "Fetches LinkedIn ad performance metrics via adAnalytics. " +
		"Use fields to request any supported metric (see input schema catalog). " +
		"Combine with linkedin_get_campaigns to map pivotValues to campaign names; use linkedin_get_campaign_groups for funnel-stage IDs. " +
		"For multi-pivot breakdowns (e.g. campaign by placement), set finder_type statistics with up to 3 pivots. " +
		"For CRM-attributed revenue (HubSpot, Salesforce, etc. connected to LinkedIn), use finder_type attributedRevenueMetrics. " +
		"For WoW reports, call twice (current week + prior week) with time_granularity ALL. " +
		"For conversion performance, first call linkedin_list_conversions to get rule names, then request externalWebsiteConversions / externalWebsitePostClickConversions / externalWebsitePostViewConversions / conversionValueInLocalCurrency pivoted by CAMPAIGN or CREATIVE. " +
		"Pagination uses start/count offsets (default auto_paginate true). Pass page_token from paging.links[rel=next] to fetch a specific page manually."
}
