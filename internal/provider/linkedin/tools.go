package linkedin

import (
	"context"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	"jumon-mcp/internal/provider/registry"
)

const platformName = "linkedin"

const (
	toolLinkedInListAdAccounts    = "linkedin_list_ad_accounts"
	toolLinkedInGetCampaignGroups = "linkedin_get_campaign_groups"
	toolLinkedInGetCampaigns      = "linkedin_get_campaigns"
	toolLinkedInGetAdAnalytics    = "linkedin_get_ad_analytics"
	toolLinkedInSearchCreatives   = "linkedin_search_creatives"
	toolLinkedInListLeadForms     = "linkedin_list_lead_forms"
	toolLinkedInGetCampaign       = "linkedin_get_campaign"
	toolLinkedInGetCreative       = "linkedin_get_creative"
	toolLinkedInListConversions   = "linkedin_list_conversions"
)

func RegisterTools(reg *registry.Registry, gatewayClient *gateway.Client) error {
	port := newLinkedInGateway(gatewayClient)
	svc := newLinkedInService(port)

	tools := []registry.ToolDefinition{
		{
			Name:               toolLinkedInListAdAccounts,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists LinkedIn ad accounts available to the authenticated user.",
			Description:        "Fetches LinkedIn ad accounts with optional filters for status, IDs, names, and pagination.",
			InputSchema:        listAdAccountsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseListAdAccountsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.listAdAccounts(ctx, userID, toolLinkedInListAdAccounts, in)
			},
		},
		{
			Name:               toolLinkedInGetCampaignGroups,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches LinkedIn campaign groups for one ad account.",
			Description:        "Lists campaign groups (TOFU/MOFU/BOFU hierarchy) with optional status, name, test, and paging filters. Use before analytics to obtain group IDs for campaign_group_ids. Pair with linkedin_get_campaigns for full account structure.",
			InputSchema:        getCampaignGroupsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseGetCampaignGroupsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getCampaignGroups(ctx, userID, toolLinkedInGetCampaignGroups, in)
			},
		},
		{
			Name:               toolLinkedInGetCampaigns,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches LinkedIn campaigns for one ad account.",
			Description:        "Fetches campaigns with optional status, campaign group, type, name, test, and paging filters. Use campaign_group_filter with IDs from linkedin_get_campaign_groups to scope by funnel stage.",
			InputSchema:        getCampaignsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseGetCampaignsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getCampaigns(ctx, userID, toolLinkedInGetCampaigns, in)
			},
		},
		{
			Name:               toolLinkedInGetAdAnalytics,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches LinkedIn ad analytics by account/campaign grouping.",
			Description:        linkedInGetAdAnalyticsToolDescription(),
			InputSchema:        getAnalyticsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseGetAnalyticsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getAnalytics(ctx, userID, toolLinkedInGetAdAnalytics, in)
			},
		},
		{
			Name:               toolLinkedInSearchCreatives,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists LinkedIn creatives for selected campaign URNs.",
			Description:        "Fetches creatives for one account and campaign IDs/URNs. Each element may include feedUrl (post permalink), previewUrl (Ad Preview API, expires ~3 hours), lead gen form fields (leadFormUrn, leadFormCtaLabel, leadFormName) when the creative has a leadgenCallToAction, and thumbnailUrl when include_asset_urls=true. Set include_preview_urls=false or include_lead_form_details=false to skip those enrichment calls.",
			InputSchema:        searchCreativesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseSearchCreativesInput(params)
				if err != nil {
					return nil, err
				}
				return svc.searchCreatives(ctx, userID, toolLinkedInSearchCreatives, in)
			},
		},
		{
			Name:               toolLinkedInListLeadForms,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists LinkedIn lead gen forms for an ad account.",
			Description:        "Fetches all lead gen forms owned by the given ad account using GET /rest/leadForms?q=owner. Returns form names, IDs, and status. Use to discover form URNs for cross-referencing leadFormUrn fields returned by linkedin_search_creatives.",
			InputSchema:        listLeadFormsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseListLeadFormsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.listLeadForms(ctx, userID, toolLinkedInListLeadForms, in)
			},
		},
		{
			Name:               toolLinkedInGetCampaign,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches a single LinkedIn campaign by ID.",
			Description:        "Returns the full campaign object for one campaign. Accepts either a numeric campaign ID or a full urn:li:sponsoredCampaign:... URN.",
			InputSchema:        getCampaignSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseGetCampaignInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getCampaign(ctx, userID, toolLinkedInGetCampaign, in)
			},
		},
		{
			Name:               toolLinkedInGetCreative,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches a single LinkedIn creative by ID with full enrichment.",
			Description:        "Returns the full creative object for one creative, enriched with feedUrl, previewUrl, lead gen form fields, and optionally thumbnailUrl (same flags as linkedin_search_creatives). Accepts either a numeric creative ID or a full urn:li:sponsoredCreative:... URN.",
			InputSchema:        getCreativeSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseGetCreativeInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getCreative(ctx, userID, toolLinkedInGetCreative, in)
			},
		},
		{
			Name:               toolLinkedInListConversions,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists LinkedIn conversion rules for an ad account.",
			Description:        "Fetches conversion tracking rules for the given account using GET /rest/conversions?q=account. Use the returned conversion IDs/URNs to filter analytics by specific conversion events in linkedin_get_ad_analytics. Set enabled_only=true to exclude paused rules.",
			InputSchema:        listConversionsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseListConversionsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.listConversions(ctx, userID, toolLinkedInListConversions, in)
			},
		},
	}

	for _, tool := range tools {
		if err := reg.Register(tool); err != nil {
			return err
		}
	}
	return nil
}

func listAdAccountsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status_filter":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"account_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"name_filter":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"reference_filter": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"test_filter":      map[string]any{"type": "boolean"},
			"page_size":        map[string]any{"type": "number"},
			"start":            map[string]any{"type": "number"},
		},
	}
}

func getCampaignsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID (without the 'urn:li:sponsoredAccount:' prefix).",
			},
			"status_filter": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Limit results to campaigns with these statuses. Valid values: ACTIVE, PAUSED, ARCHIVED, COMPLETED, CANCELED, DRAFT, PENDING_DELETION, REMOVED. Omit to return all statuses.",
			},
			"campaign_group_filter": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Numeric campaign group IDs or full URNs (urn:li:sponsoredCampaignGroup:<id>) to scope results. Obtain IDs from linkedin_get_campaign_groups. Applied server-side after fetching campaigns (auto_paginate is forced when this filter is set).",
			},
			"type_filter": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Campaign types to include, e.g. TEXT_AD, SPONSORED_UPDATES, SPONSORED_INMAILS.",
			},
			"name_filter": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Exact campaign names to match.",
			},
			"test_filter": map[string]any{
				"type":        "boolean",
				"description": "Set true to return only test campaigns; false for live campaigns.",
			},
			"sort_order": map[string]any{
				"type":        "string",
				"enum":        []string{"ASCENDING", "DESCENDING"},
				"description": "Sort direction by campaign ID. DESCENDING (default) returns newest campaigns first.",
			},
			"auto_paginate": map[string]any{
				"type":        "boolean",
				"description": "When true (default), automatically follows nextPageToken to fetch ALL campaigns across multiple pages. Set to false to get only one page. Always use true when building reports to avoid missing campaigns.",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Results per page (default 100, max 100). Only relevant when auto_paginate is false.",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Pagination cursor from a previous response. Setting this disables auto_paginate.",
			},
		},
	}
}

func getCampaignGroupsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID (without the 'urn:li:sponsoredAccount:' prefix).",
			},
			"status_filter": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Limit results to campaign groups with these statuses. Valid values: ACTIVE, PAUSED, ARCHIVED, DRAFT, CANCELED. Omit to return all statuses.",
			},
			"name_filter": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Exact campaign group names to match.",
			},
			"test_filter": map[string]any{
				"type":        "boolean",
				"description": "Set true to return only test campaign groups; false for live groups.",
			},
			"sort_order": map[string]any{
				"type":        "string",
				"enum":        []string{"ASCENDING", "DESCENDING"},
				"description": "Sort direction by campaign group ID. DESCENDING (default) returns newest groups first.",
			},
			"auto_paginate": map[string]any{
				"type":        "boolean",
				"description": "When true (default), automatically follows nextPageToken to fetch ALL campaign groups across multiple pages. Set to false to get only one page.",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Results per page (default 100, max 100). Only relevant when auto_paginate is false.",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Pagination cursor from a previous response. Setting this disables auto_paginate.",
			},
		},
	}
}

func getAnalyticsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id", "date_range_start", "pivots"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID.",
			},
			"date_range_start": map[string]any{
				"type":        "string",
				"description": "Inclusive start date in YYYY-MM-DD format.",
			},
			"date_range_end": map[string]any{
				"type":        "string",
				"description": "Inclusive end date in YYYY-MM-DD format. Defaults to today when omitted.",
			},
			"pivots": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
				"description": "Grouping dimension(s). finder_type analytics (default) uses only the first pivot (singular pivot=). " +
					"finder_type statistics accepts 1–3 pivots (pivots=List(...)) for multi-dimensional breakdowns, e.g. CAMPAIGN + PLACEMENT_NAME. " +
					"finder_type attributedRevenueMetrics accepts 1–3 pivots: ACCOUNT, CAMPAIGN_GROUP, or CAMPAIGN only. " +
					"Structure: ACCOUNT, CAMPAIGN_GROUP, CAMPAIGN, CREATIVE, SHARE, COMPANY, CONVERSION. " +
					"Demographics: MEMBER_COMPANY_SIZE, MEMBER_INDUSTRY, MEMBER_SENIORITY, MEMBER_JOB_TITLE, MEMBER_JOB_FUNCTION, MEMBER_COUNTRY_V2, MEMBER_REGION_V2, MEMBER_COMPANY. " +
					"Other: SERVING_LOCATION, PLACEMENT_NAME, OBJECTIVE_TYPE, CARD_INDEX, IMPRESSION_DEVICE_TYPE, EVENT_STAGE. " +
					"Reach and averageFrequency are unavailable on MEMBER_* pivots (top-100 values per creative per day; min 3 events per value). Each row includes pivotValues; MEMBER_INDUSTRY responses also include pivotLabels (human-readable industry names).",
				"examples": [][]string{
					{"CAMPAIGN"},
					{"CAMPAIGN_GROUP"},
					{"ACCOUNT"},
					{"CREATIVE"},
					{"CAMPAIGN", "PLACEMENT_NAME"},
					{"MEMBER_INDUSTRY"},
					{"MEMBER_SENIORITY"},
					{"MEMBER_JOB_TITLE"},
				},
			},
			"time_granularity": map[string]any{
				"type":        "string",
				"enum":        []string{"ALL", "DAILY", "MONTHLY", "YEARLY"},
				"description": "Time bucketing for finder_type analytics or statistics. Use ALL for a single aggregate row per pivot entity, DAILY for day-by-day breakdown. Not supported for attributedRevenueMetrics.",
			},
			"finder_type": map[string]any{
				"type": "string",
				"enum": []string{"analytics", "statistics", "attributedRevenueMetrics"},
				"description": "LinkedIn finder to use. Default 'analytics' for delivery and performance metrics (single pivot). " +
					"Use 'statistics' for up to 3 pivots in one request (e.g. CAMPAIGN + PLACEMENT_NAME). " +
					"Use 'attributedRevenueMetrics' for CRM-attributed revenue (requires CRM connected to LinkedIn Campaign Manager; date range 30–366 days; no time_granularity; request revenueAttributionMetrics not top-level revenue field names; pivots ACCOUNT, CAMPAIGN_GROUP, or CAMPAIGN; openOpportunities/opportunityAmountInUsd only when date_range_end is today UTC).",
			},
			"campaign_group_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Numeric campaign group IDs to scope results. Obtain IDs from linkedin_get_campaign_groups.",
			},
			"campaign_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Numeric campaign IDs to scope results. Use this to query specific campaigns individually when the full account list is large.",
			},
			"creative_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Numeric creative IDs to scope results.",
			},
			"fields": analyticsFieldsSchemaProperty(),
			"sort_by_field": map[string]any{
				"type":        "string",
				"description": "Metric field name to sort results by, e.g. spend.",
			},
			"sort_by_order": map[string]any{
				"type":        "string",
				"enum":        []string{"ASCENDING", "DESCENDING"},
				"description": "Sort direction (default DESCENDING).",
			},
		},
	}
}

func searchCreativesSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id"},
		"properties": map[string]any{
			"account_id": map[string]any{"type": "string"},
			"campaign_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Numeric campaign IDs or full campaign URNs.",
			},
			"campaign_urns": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"sort_order":    map[string]any{"type": "string", "enum": []string{"ASCENDING", "DESCENDING"}},
			"page_size":     map[string]any{"type": "number"},
			"page_token":    map[string]any{"type": "string"},
			"auto_paginate": map[string]any{
				"type":        "boolean",
				"description": "When true (default), fetches all pages of creatives for the campaign filter. Disabled when page_token is set.",
			},
			"include_preview_urls": map[string]any{
				"type":        "boolean",
				"description": "When true (default), fetches previewUrl per creative from LinkedIn Ad Preview API (one extra API call per creative; preview links expire ~3 hours). Set false to return only feedUrl when available.",
			},
			"include_lead_form_details": map[string]any{
				"type":        "boolean",
				"description": "When true (default), resolves lead gen form name and CTA label for creatives with a leadgenCallToAction (one batch API call for all unique forms). Set false to skip. Adds leadFormUrn, leadFormCtaLabel, and leadFormName fields when available.",
			},
			"include_asset_urls": map[string]any{
				"type":        "boolean",
				"description": "When true, attempts to fetch thumbnailUrl per creative by resolving the post media chain (posts → images/videos API). Default false. NOTE: As of the current r_ads OAuth scope, the posts endpoint returns an access error and thumbnailUrl will not be populated. This flag is reserved for when r_organization_social or an equivalent scope is granted.",
			},
		},
	}
}

func listLeadFormsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID.",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Maximum number of forms to return per page.",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Pagination cursor (numeric start offset) from a previous response.",
			},
		},
	}
}

func getCampaignSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id", "campaign_id"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID.",
			},
			"campaign_id": map[string]any{
				"type":        "string",
				"description": "Numeric campaign ID or full urn:li:sponsoredCampaign:... URN.",
			},
		},
	}
}

func getCreativeSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id", "creative_id"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID.",
			},
			"creative_id": map[string]any{
				"type":        "string",
				"description": "Numeric creative ID or full urn:li:sponsoredCreative:... URN.",
			},
			"include_preview_urls": map[string]any{
				"type":        "boolean",
				"description": "When true (default), fetches previewUrl from Ad Preview API.",
			},
			"include_lead_form_details": map[string]any{
				"type":        "boolean",
				"description": "When true (default), resolves lead gen form name and CTA. Adds leadFormUrn, leadFormCtaLabel, leadFormName when available.",
			},
			"include_asset_urls": map[string]any{
				"type":        "boolean",
				"description": "When true, attempts to fetch thumbnailUrl via posts/images API. Default false. Currently degrades silently under r_ads scope (posts endpoint not accessible).",
			},
		},
	}
}

func listConversionsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"account_id"},
		"properties": map[string]any{
			"account_id": map[string]any{
				"type":        "string",
				"description": "Numeric LinkedIn ad account ID.",
			},
			"enabled_only": map[string]any{
				"type":        "boolean",
				"description": "When true, returns only enabled (active) conversion rules. Default false returns all rules.",
			},
		},
	}
}
