package meta

import (
	"context"
	"strings"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	"jumon-mcp/internal/provider/registry"
)

const (
	toolMetaListAdAccounts       = "meta_list_ad_accounts"
	toolMetaGetAdAccount         = "meta_get_ad_account"
	toolMetaListCampaigns        = "meta_list_campaigns"
	toolMetaGetCampaign          = "meta_get_campaign"
	toolMetaListAdSets           = "meta_list_ad_sets"
	toolMetaListAds              = "meta_list_ads"
	toolMetaGetAdAccountInsights = "meta_get_ad_account_insights"
	toolMetaGetCampaignInsights  = "meta_get_campaign_insights"
	toolMetaSearchAdEntities     = "meta_search_ad_entities"
	toolMetaGetFieldContext      = "meta_get_field_context"
)

func RegisterTools(reg *registry.Registry, gatewayClient *gateway.Client) error {
	port := newMetaGateway(gatewayClient)
	svc := newService(port)

	tools := []registry.ToolDefinition{
		{
			Name:               toolMetaListAdAccounts,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Meta ad accounts for the connected user.",
			Description:        "Calls GET /me?fields=adaccounts{...}. Start here to obtain act_id values.",
			InputSchema:        listAdAccountsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.listAdAccounts(ctx, toolMetaListAdAccounts, userID)
			},
		},
		{
			Name:               toolMetaGetAdAccount,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Gets details for one Meta ad account.",
			Description:        "Calls GET /{act_id} with curated default fields (currency, status, spend). " + docActID,
			InputSchema:        getAdAccountSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := getAdAccountInput{
					actID:  strings.TrimSpace(toString(params["act_id"])),
					fields: toStringSlice(params["fields"]),
				}
				return svc.getAdAccount(ctx, toolMetaGetAdAccount, userID, in)
			},
		},
		{
			Name:               toolMetaListCampaigns,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists campaigns in a Meta ad account.",
			Description:        "Calls GET /{act_id}/campaigns with optional effective_status filters. " + docAutoPaginate,
			InputSchema:        listCampaignsSchema(),
			RequiresConnection: true,
			Execute:            listCampaignsExecutor(svc, toolMetaListCampaigns),
		},
		{
			Name:               toolMetaGetCampaign,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Gets one Meta campaign by id.",
			Description:        "Calls GET /{campaign_id}.",
			InputSchema:        getCampaignSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := getCampaignInput{
					campaignID: strings.TrimSpace(toString(params["campaign_id"])),
					fields:     toStringSlice(params["fields"]),
				}
				return svc.getCampaign(ctx, toolMetaGetCampaign, userID, in)
			},
		},
		{
			Name:               toolMetaListAdSets,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists ad sets in a Meta ad account.",
			Description:        "Calls GET /{act_id}/adsets. Optional campaign_id filter. " + docAutoPaginate,
			InputSchema:        listAdSetsSchema(),
			RequiresConnection: true,
			Execute:            listAdSetsExecutor(svc, toolMetaListAdSets),
		},
		{
			Name:               toolMetaListAds,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists ads in a Meta ad account.",
			Description:        "Calls GET /{act_id}/ads. Optional campaign_id or adset_id filters. " + docAutoPaginate,
			InputSchema:        listAdsSchema(),
			RequiresConnection: true,
			Execute:            listAdsExecutor(svc, toolMetaListAds),
		},
		{
			Name:               toolMetaGetAdAccountInsights,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Account-level performance insights.",
			Description:        docInsightsTimePrecedence + " " + docPublisherPlatform + " " + docNoStandaloneActions,
			InputSchema:        adAccountInsightsSchema(),
			RequiresConnection: true,
			Execute:            adAccountInsightsExecutor(svc, toolMetaGetAdAccountInsights),
		},
		{
			Name:               toolMetaGetCampaignInsights,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Campaign-level performance insights.",
			Description:        docInsightsTimePrecedence + " " + docPublisherPlatform,
			InputSchema:        campaignInsightsSchema(),
			RequiresConnection: true,
			Execute:            campaignInsightsExecutor(svc, toolMetaGetCampaignInsights),
		},
		{
			Name:               toolMetaSearchAdEntities,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Unified insights report at account/campaign/adset/ad level.",
			Description:        docSearchEntitiesPreferred + " " + docInsightsTimePrecedence + " Call meta_get_field_context before filtering/sorting.",
			InputSchema:        searchAdEntitiesSchema(),
			RequiresConnection: true,
			Execute:            searchAdEntitiesExecutor(svc, toolMetaSearchAdEntities),
		},
		{
			Name:               toolMetaGetFieldContext,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Field metadata for Insights filtering and sorting.",
			Description:        "Returns embedded catalog of common fields (filterable, sortable, levels). Use before meta_search_ad_entities filtering.",
			InputSchema:        fieldContextSchema(),
			RequiresConnection: false,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				_ = ctx
				_ = userID
				return buildFieldContextResponse(fieldContextInput{
					fieldNames: toStringSlice(params["field_names"]),
					level:      strings.TrimSpace(toString(params["level"])),
				}), nil
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

func listCampaignsExecutor(svc *service, mcpTool string) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listCampaignsInput{listPaginationInput: parseListPagination(params, defaultCampaignListFields)}
		actID := strings.TrimSpace(toString(params["act_id"]))
		return svc.listCampaigns(ctx, mcpTool, userID, actID, in)
	}
}

func listAdSetsExecutor(svc *service, mcpTool string) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listAdSetsInput{
			listPaginationInput: parseListPagination(params, defaultAdSetListFields),
			campaignID:        strings.TrimSpace(toString(params["campaign_id"])),
		}
		actID := strings.TrimSpace(toString(params["act_id"]))
		return svc.listAdSets(ctx, mcpTool, userID, actID, in)
	}
}

func listAdsExecutor(svc *service, mcpTool string) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listAdsInput{
			listPaginationInput: parseListPagination(params, defaultAdListFields),
			campaignID:        strings.TrimSpace(toString(params["campaign_id"])),
			adSetID:           strings.TrimSpace(toString(params["adset_id"])),
		}
		actID := strings.TrimSpace(toString(params["act_id"]))
		return svc.listAds(ctx, mcpTool, userID, actID, in)
	}
}

func adAccountInsightsExecutor(svc *service, mcpTool string) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in, err := parseInsightsInput(params, defaultInsightsFields)
		if err != nil {
			return nil, err
		}
		actID := strings.TrimSpace(toString(params["act_id"]))
		return svc.getAdAccountInsights(ctx, mcpTool, userID, actID, in)
	}
}

func campaignInsightsExecutor(svc *service, mcpTool string) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in, err := parseInsightsInput(params, defaultInsightsFields)
		if err != nil {
			return nil, err
		}
		campaignID := strings.TrimSpace(toString(params["campaign_id"]))
		return svc.getCampaignInsights(ctx, mcpTool, userID, campaignID, in)
	}
}

func searchAdEntitiesExecutor(svc *service, mcpTool string) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		insights, err := parseInsightsInput(params, defaultSearchEntitiesFields)
		if err != nil {
			return nil, err
		}
		in := searchAdEntitiesInput{
			actID:         strings.TrimSpace(toString(params["act_id"])),
			insightsInput: insights,
		}
		return svc.searchAdEntities(ctx, mcpTool, userID, in)
	}
}
