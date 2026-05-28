package googleads

import (
	"context"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	"jumon-mcp/internal/provider/registry"
)

const platformName = "google"

const (
	toolGoogleListAdAccounts             = "google_list_ad_accounts"
	toolGoogleResolveCustomer            = "google_resolve_customer"
	toolGoogleListClientAccountsUnderMgr = "google_list_client_accounts_under_manager"
	toolGoogleSearchCampaigns            = "google_search_campaigns"
	toolGoogleSearchAdGroups             = "google_search_ad_groups"
	toolGoogleSearchAds                  = "google_search_ads"
	toolGoogleSearchKeywords             = "google_search_keywords"
	toolGoogleSearchSearchTerms          = "google_search_search_terms"
	toolGoogleSearchPmaxSearchTerms      = "google_search_pmax_search_terms"
	toolGoogleListConversionActions      = "google_list_conversion_actions"
	toolGoogleSearchConversionPerf       = "google_search_conversion_performance"
	toolGoogleListOfflineConvUploads   = "google_list_offline_conversion_upload_summaries"
	toolGoogleGetResourceMetadata        = "google_get_resource_metadata"
	toolGoogleSearchGAQL                 = "google_search_gaql"
)

type Config struct {
	APIVersion            string
	MaxAccessibleAccounts int
	MaxManagerScan        int
}

func RegisterTools(reg *registry.Registry, gatewayClient *gateway.Client, config Config) error {
	if config.APIVersion == "" {
		config.APIVersion = "v22"
	}

	svc := newGoogleService(gatewayClient, config)

	tools := []registry.ToolDefinition{
		{
			Name:               toolGoogleListAdAccounts,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists accessible Google Ads accounts with names, manager flag, currency, and timezone.",
			Description:        listAdAccountsDescription(),
			InputSchema:        map[string]any{"type": "object"},
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, _ map[string]any) (any, error) {
				return svc.listAdAccounts(ctx, userID, toolGoogleListAdAccounts)
			},
		},
		{
			Name:               toolGoogleResolveCustomer,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Find customer_id and login_customer_id by account name when the user does not provide a CID.",
			Description:        resolveCustomerDescription(),
			InputSchema:        resolveCustomerSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseResolveCustomerInput(params)
				if err != nil {
					return nil, err
				}
				return svc.resolveCustomer(ctx, userID, toolGoogleResolveCustomer, in)
			},
		},
		{
			Name:               toolGoogleListClientAccountsUnderMgr,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists client accounts under one manager (MCC); optional client name filter.",
			Description:        listClientAccountsDescription(),
			InputSchema:        listClientAccountsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseListClientAccountsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.listClientAccounts(ctx, userID, toolGoogleListClientAccountsUnderMgr, in)
			},
		},
		{
			Name:               toolGoogleSearchCampaigns,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Google Ads campaigns and core metrics.",
			Description:        "Runs a GAQL campaign query with optional status/name/date filters. " + googleAdsWorkflowHint(),
			InputSchema:        searchCampaignsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseSearchCampaignsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.searchCampaigns(ctx, userID, toolGoogleSearchCampaigns, in)
			},
		},
		{
			Name:               toolGoogleSearchAdGroups,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Google Ads ad groups and key metrics.",
			Description:        "Runs a GAQL ad_group query with optional campaign/status/date filters. " + googleAdsWorkflowHint(),
			InputSchema:        searchAdGroupsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseSearchAdGroupsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.searchAdGroups(ctx, userID, toolGoogleSearchAdGroups, in)
			},
		},
		{
			Name:               toolGoogleSearchAds,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Google Ads ad-level entities and metrics.",
			Description:        "Runs a GAQL ad_group_ad query with optional campaign/adgroup/status/date filters. " + googleAdsWorkflowHint(),
			InputSchema:        searchAdsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseSearchAdsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.searchAds(ctx, userID, toolGoogleSearchAds, in)
			},
		},
		{
			Name:               toolGoogleSearchKeywords,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Keyword-level performance (keyword_view) with optional campaign/ad group and date filters.",
			Description:        searchKeywordsDescription(),
			InputSchema:        searchKeywordsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseReportFilters(params)
				if err != nil {
					return nil, err
				}
				return svc.searchKeywords(ctx, userID, toolGoogleSearchKeywords, in)
			},
		},
		{
			Name:               toolGoogleSearchSearchTerms,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Search query / search term report (search_term_view) with optional date range.",
			Description:        searchSearchTermsDescription(),
			InputSchema:        searchSearchTermsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseReportFilters(params)
				if err != nil {
					return nil, err
				}
				return svc.searchSearchTerms(ctx, userID, toolGoogleSearchSearchTerms, in)
			},
		},
		{
			Name:               toolGoogleSearchPmaxSearchTerms,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Performance Max search term report (campaign_search_term_view) with optional date range.",
			Description:        searchPmaxSearchTermsDescription(),
			InputSchema:        searchPmaxSearchTermsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseReportFilters(params)
				if err != nil {
					return nil, err
				}
				return svc.searchPmaxSearchTerms(ctx, userID, toolGoogleSearchPmaxSearchTerms, in)
			},
		},
		{
			Name:               toolGoogleListConversionActions,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "List conversion action definitions (names, types, status) for an account.",
			Description:        listConversionActionsDescription(),
			InputSchema:        listConversionActionsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseReportFilters(params)
				if err != nil {
					return nil, err
				}
				return svc.listConversionActions(ctx, userID, toolGoogleListConversionActions, in)
			},
		},
		{
			Name:               toolGoogleSearchConversionPerf,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Conversion metrics by campaign and conversion action over a date range.",
			Description:        searchConversionPerformanceDescription(),
			InputSchema:        searchConversionPerformanceSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseConversionPerformanceFilters(params)
				if err != nil {
					return nil, err
				}
				return svc.searchConversionPerformance(ctx, userID, toolGoogleSearchConversionPerf, in)
			},
		},
		{
			Name:               toolGoogleListOfflineConvUploads,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Offline conversion upload health by conversion action.",
			Description:        listOfflineConversionUploadSummariesDescription(),
			InputSchema:        listOfflineConversionUploadSummariesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseReportFilters(params)
				if err != nil {
					return nil, err
				}
				return svc.listOfflineConversionUploadSummaries(ctx, userID, toolGoogleListOfflineConvUploads, in)
			},
		},
		{
			Name:               toolGoogleGetResourceMetadata,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Discover selectable GAQL fields for a Google Ads resource.",
			Description:        getResourceMetadataDescription(),
			InputSchema:        getResourceMetadataSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				resourceName, err := parseResourceMetadataInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getResourceMetadata(ctx, userID, toolGoogleGetResourceMetadata, resourceName)
			},
		},
		{
			Name:               toolGoogleSearchGAQL,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Run a validated GAQL query when no curated Google tool fits.",
			Description:        searchGAQLDescription(),
			InputSchema:        searchGAQLSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseGAQLSearchInput(params)
				if err != nil {
					return nil, err
				}
				return svc.searchGAQL(ctx, userID, toolGoogleSearchGAQL, in)
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
