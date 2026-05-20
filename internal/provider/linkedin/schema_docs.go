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
	b.WriteString("When pivots are set, pivotValues, approximateMemberReach, and impressions are auto-included if missing. ")
	b.WriteString("Response may include averageFrequency (derived: impressions / approximateMemberReach, Campaign Manager Average frequency).\n\n")

	for _, cat := range linkedInAnalyticsFieldCatalog {
		b.WriteString(cat.category)
		b.WriteString(": ")
		b.WriteString(strings.Join(cat.fields, ", "))
		b.WriteString(".\n")
	}

	b.WriteString("\nConstraints: reach metrics need date range <= 92 days and non-MEMBER_* pivots; ")
	b.WriteString("use time_granularity ALL for weekly totals (do not sum daily reach). ")
	b.WriteString("For CRM revenue metrics use finder_type attributedRevenueMetrics. ")
	b.WriteString("Viral variants exist (viralImpressions, viralVideoViews, etc.) — prefix viral to the paid metric name.")

	return b.String()
}

func linkedInGetAdAnalyticsToolDescription() string {
	return "Fetches LinkedIn ad performance metrics via adAnalytics. " +
		"Use fields to request any supported metric (see input schema catalog). " +
		"Combine with linkedin_get_campaigns to map pivotValues to campaign names; use linkedin_get_campaign_groups for funnel-stage IDs. " +
		"For multi-pivot breakdowns (e.g. campaign by placement), set finder_type statistics with up to 3 pivots. " +
		"For WoW reports, call twice (current week + prior week) with time_granularity ALL."
}
