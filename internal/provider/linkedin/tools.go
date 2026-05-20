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
			Description:        "Fetches creatives via LinkedIn criteria finder for one account and one or more campaign IDs/URNs.",
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
					"Structure: ACCOUNT, CAMPAIGN_GROUP, CAMPAIGN, CREATIVE, SHARE, COMPANY, CONVERSION. " +
					"Demographics: MEMBER_COMPANY_SIZE, MEMBER_INDUSTRY, MEMBER_SENIORITY, MEMBER_JOB_TITLE, MEMBER_JOB_FUNCTION, MEMBER_COUNTRY_V2, MEMBER_REGION_V2. " +
					"Other: SERVING_LOCATION, PLACEMENT_NAME, OBJECTIVE_TYPE, CARD_INDEX, IMPRESSION_DEVICE_TYPE, EVENT_STAGE. " +
					"Reach metrics require non-MEMBER_* pivots. Each row includes pivotValues (entity URN).",
				"examples": [][]string{
					{"CAMPAIGN"},
					{"CAMPAIGN_GROUP"},
					{"ACCOUNT"},
					{"CREATIVE"},
					{"CAMPAIGN", "PLACEMENT_NAME"},
				},
			},
			"time_granularity": map[string]any{
				"type":        "string",
				"enum":        []string{"ALL", "DAILY", "MONTHLY", "YEARLY"},
				"description": "Time bucketing. Use ALL for a single aggregate row per pivot entity, DAILY for day-by-day breakdown.",
			},
			"finder_type": map[string]any{
				"type":        "string",
				"enum":        []string{"analytics", "statistics", "attributedRevenueMetrics"},
				"description": "LinkedIn finder to use. Default 'analytics' for delivery and performance metrics (single pivot). Use 'statistics' for up to 3 pivots in one request (e.g. CAMPAIGN + PLACEMENT_NAME). Use 'attributedRevenueMetrics' for CRM-attributed revenue fields.",
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
			"account_id":    map[string]any{"type": "string"},
			"campaign_ids":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"campaign_urns": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"sort_order":    map[string]any{"type": "string", "enum": []string{"ASCENDING", "DESCENDING"}},
			"page_size":     map[string]any{"type": "number"},
			"page_token":    map[string]any{"type": "string"},
		},
	}
}
