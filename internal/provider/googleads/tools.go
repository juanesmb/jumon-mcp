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
	toolGoogleListClientAccountsUnderMgr = "google_list_client_accounts_under_manager"
	toolGoogleSearchCampaigns            = "google_search_campaigns"
	toolGoogleSearchAdGroups             = "google_search_ad_groups"
	toolGoogleSearchAds                  = "google_search_ads"
)

type Config struct {
	APIVersion string
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
			Summary:            "Lists Google Ads customer IDs accessible to the connected user.",
			Description:        "Calls customers:listAccessibleCustomers and returns accessible customer resource names.",
			InputSchema:        map[string]any{"type": "object"},
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, _ map[string]any) (any, error) {
				return svc.listAdAccounts(ctx, userID, toolGoogleListAdAccounts)
			},
		},
		{
			Name:               toolGoogleListClientAccountsUnderMgr,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists non-manager client accounts under one manager account.",
			Description:        "Runs a GAQL customer_client query under an MCC manager account.",
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
			Description:        "Runs a GAQL campaign query with optional status/name/date filters.",
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
			Description:        "Runs a GAQL ad_group query with optional campaign/status/date filters.",
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
			Description:        "Runs a GAQL ad_group_ad query with optional campaign/adgroup/status/date filters.",
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
	}

	for _, tool := range tools {
		if err := reg.Register(tool); err != nil {
			return err
		}
	}
	return nil
}

func listClientAccountsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"customer_id"},
		"properties": map[string]any{
			"customer_id":       map[string]any{"type": "string"},
			"login_customer_id": map[string]any{"type": "string"},
		},
	}
}

func searchCampaignsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"customer_id"},
		"properties": map[string]any{
			"customer_id":            map[string]any{"type": "string"},
			"login_customer_id":      map[string]any{"type": "string"},
			"campaign_ids":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"campaign_name_contains": map[string]any{"type": "string"},
			"statuses":               map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"date_range_start":       map[string]any{"type": "string"},
			"date_range_end":         map[string]any{"type": "string"},
		},
	}
}

func searchAdGroupsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"customer_id"},
		"properties": map[string]any{
			"customer_id":       map[string]any{"type": "string"},
			"login_customer_id": map[string]any{"type": "string"},
			"ad_group_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"campaign_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"statuses":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"date_range_start":  map[string]any{"type": "string"},
			"date_range_end":    map[string]any{"type": "string"},
		},
	}
}

func searchAdsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"customer_id"},
		"properties": map[string]any{
			"customer_id":       map[string]any{"type": "string"},
			"login_customer_id": map[string]any{"type": "string"},
			"ad_ids":            map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"ad_group_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"campaign_ids":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"statuses":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"date_range_start":  map[string]any{"type": "string"},
			"date_range_end":    map[string]any{"type": "string"},
		},
	}
}
