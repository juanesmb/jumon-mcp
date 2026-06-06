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
	toolMetaGetAdSet             = "meta_get_ad_set"
	toolMetaGetAd                = "meta_get_ad"
	toolMetaGetDeliveryErrors    = "meta_get_delivery_errors"
	toolMetaListAccountPages     = "meta_list_account_pages"
	toolMetaListCreatives        = "meta_list_creatives"
	toolMetaGetCreative          = "meta_get_creative"
	toolMetaGetAdImages          = "meta_get_ad_images"
	toolMetaGetAdVideos          = "meta_get_ad_videos"
	toolMetaGetAdPreview         = "meta_get_ad_preview"
	toolMetaSearchInterests      = "meta_search_interests"
	toolMetaSearchGeoLocations   = "meta_search_geo_locations"
	toolMetaEstimateAudienceSize = "meta_estimate_audience_size"
	toolMetaListCustomAudiences  = "meta_list_custom_audiences"
	toolMetaGetCustomAudience    = "meta_get_custom_audience"
	toolMetaListCustomAudienceAdSets = "meta_list_custom_audience_ad_sets"
	toolMetaGetOpportunityScore       = "meta_get_opportunity_score"
	toolMetaListCustomConversions     = "meta_list_custom_conversions"
	toolMetaListDatasets              = "meta_list_datasets"
	toolMetaGetDataset                = "meta_get_dataset"
	toolMetaListCreativeAds           = "meta_list_creative_ads"
	toolMetaGetAccountActivities      = "meta_get_account_activities"
	toolMetaSearchBehaviors           = "meta_search_behaviors"
	toolMetaSearchDemographics        = "meta_search_demographics"
	toolMetaGetInterestSuggestions    = "meta_get_interest_suggestions"
	toolMetaGetDatasetStats           = "meta_get_dataset_stats"
	toolMetaGetDatasetQuality         = "meta_get_dataset_quality"
	toolMetaGetAdSetActivities        = "meta_get_ad_set_activities"
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
			Description:        "Calls GET /me?fields=adaccounts{...}. Start here to obtain act_id values. " + docAccountListNote,
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
			Description:        docSearchEntitiesPreferred + " " + docInsightsLevels + " " + docInsightsTimePrecedence + " Call meta_get_field_context before filtering/sorting.",
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
		{
			Name:               toolMetaGetAdSet,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Gets one Meta ad set by id.",
			Description:        "Calls GET /{adset_id}.",
			InputSchema:        getAdSetSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := getAdSetInput{
					adSetID: strings.TrimSpace(toString(params["adset_id"])),
					fields:  toStringSlice(params["fields"]),
				}
				return svc.getAdSet(ctx, toolMetaGetAdSet, userID, in)
			},
		},
		{
			Name:               toolMetaGetAd,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Gets one Meta ad by id.",
			Description:        "Calls GET /{ad_id}.",
			InputSchema:        getAdSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := getAdInput{
					adID:   strings.TrimSpace(toString(params["ad_id"])),
					fields: toStringSlice(params["fields"]),
				}
				return svc.getAd(ctx, toolMetaGetAd, userID, in)
			},
		},
		{
			Name:               toolMetaGetDeliveryErrors,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches delivery-blocking errors for campaigns, ad sets, or ads.",
			Description:        "Batch GET per entity_id with failed_delivery_checks (ads) and issues_info (campaigns/ad sets). Does not cover pacing or account disablement.",
			InputSchema:        deliveryErrorsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseDeliveryErrorsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getDeliveryErrors(ctx, toolMetaGetDeliveryErrors, userID, in)
			},
		},
		{
			Name:               toolMetaListAccountPages,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Facebook Pages available for advertising.",
			Description:        "Without act_id: GET /me/accounts. With act_id: GET /{act_id}/promote_pages. " + docAccountPagesScope + " " + docLeadGenTOS + " " + docAutoPaginate,
			InputSchema:        listAccountPagesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.listAccountPages(ctx, toolMetaListAccountPages, userID, parseListAccountPagesInput(params))
			},
		},
		{
			Name:               toolMetaListCreatives,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists ad creatives in a Meta ad account.",
			Description:        "Calls GET /{act_id}/adcreatives with optional effective_status and filtering. " + docAutoPaginate,
			InputSchema:        listCreativesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseListCreativesInput(params)
				if err != nil {
					return nil, err
				}
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.listCreatives(ctx, toolMetaListCreatives, userID, actID, in)
			},
		},
		{
			Name:               toolMetaGetCreative,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Gets one Meta ad creative by id.",
			Description:        "Calls GET /{creative_id} with optional thumbnail dimensions.",
			InputSchema:        getCreativeSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.getCreative(ctx, toolMetaGetCreative, userID, parseGetCreativeInput(params))
			},
		},
		{
			Name:               toolMetaGetAdImages,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists ad images in a Meta ad account.",
			Description:        "Calls GET /{act_id}/adimages. Filter by hashes, name, or minimum dimensions. " + docAutoPaginate,
			InputSchema:        adImagesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := parseAdImagesInput(params)
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.getAdImages(ctx, toolMetaGetAdImages, userID, actID, in)
			},
		},
		{
			Name:               toolMetaGetAdVideos,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists ad videos in a Meta ad account.",
			Description:        "Calls GET /{act_id}/advideos. Optional video_ids filter. " + docAutoPaginate,
			InputSchema:        adVideosSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := parseAdVideosInput(params)
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.getAdVideos(ctx, toolMetaGetAdVideos, userID, actID, in)
			},
		},
		{
			Name:               toolMetaGetAdPreview,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Generates ad preview HTML for a placement format.",
			Description:        "Calls GET /{ad_id}/previews. Pass ad_format for a specific placement (e.g. INSTAGRAM_STANDARD).",
			InputSchema:        adPreviewSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.getAdPreview(ctx, toolMetaGetAdPreview, userID, parseAdPreviewInput(params))
			},
		},
		{
			Name:               toolMetaSearchInterests,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Meta interest targeting options.",
			Description:        "Calls GET /search?type=adinterest&q=.... Returns interest ids for targeting.flexible_spec.",
			InputSchema:        searchInterestsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.searchInterests(ctx, toolMetaSearchInterests, userID, parseSearchInterestsInput(params))
			},
		},
		{
			Name:               toolMetaSearchGeoLocations,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Meta geo targeting locations.",
			Description:        "Calls GET /search?type=adgeolocation&q=.... Returns location keys for targeting.geo_locations.",
			InputSchema:        searchGeoLocationsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.searchGeoLocations(ctx, toolMetaSearchGeoLocations, userID, parseSearchGeoInput(params))
			},
		},
		{
			Name:               toolMetaEstimateAudienceSize,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Estimates audience size for a targeting spec.",
			Description:        "Calls GET /{act_id}/delivery_estimate with targeting_spec and optimization_goal (default REACH).",
			InputSchema:        estimateAudienceSizeSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseEstimateAudienceInput(params)
				if err != nil {
					return nil, err
				}
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.estimateAudienceSize(ctx, toolMetaEstimateAudienceSize, userID, actID, in)
			},
		},
		{
			Name:               toolMetaListCustomAudiences,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists custom audiences for a Meta ad account.",
			Description:        "Calls GET /{act_id}/customaudiences. Optional subtype_filter. " + docAutoPaginate,
			InputSchema:        listCustomAudiencesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := parseListCustomAudiencesInput(params)
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.listCustomAudiences(ctx, toolMetaListCustomAudiences, userID, actID, in)
			},
		},
		{
			Name:               toolMetaGetCustomAudience,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Gets one custom audience by id.",
			Description:        "Calls GET /{custom_audience_id} with size, delivery status, and subtype fields.",
			InputSchema:        getCustomAudienceSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.getCustomAudience(ctx, toolMetaGetCustomAudience, userID, parseGetCustomAudienceInput(params))
			},
		},
		{
			Name:               toolMetaListCustomAudienceAdSets,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists ad sets targeting a custom audience.",
			Description:        "Calls GET /{custom_audience_id}/adsets. Use before deleting an audience to see impacted ad sets. " + docAutoPaginate,
			InputSchema:        listCustomAudienceAdSetsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.listCustomAudienceAdSets(ctx, toolMetaListCustomAudienceAdSets, userID, parseListCustomAudienceAdSetsInput(params))
			},
		},
		{
			Name:               toolMetaGetOpportunityScore,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Account-level opportunity score and recommendations.",
			Description:        "Calls GET /{act_id}/recommendations. Score is account-level only — not per campaign or ad. Refer to lift as points, not impact.",
			InputSchema:        opportunityScoreSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.getOpportunityScore(ctx, toolMetaGetOpportunityScore, userID, actID)
			},
		},
		{
			Name:               toolMetaListCustomConversions,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists custom conversion events on a Meta ad account.",
			Description:        docMeasurementWorkflow + " Calls GET /{act_id}/customconversions.",
			InputSchema:        listCustomConversionsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := parseListCustomConversionsInput(params)
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.listCustomConversions(ctx, toolMetaListCustomConversions, userID, actID, in)
			},
		},
		{
			Name:               toolMetaListDatasets,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists pixel/dataset nodes on a Meta ad account.",
			Description:        "Calls GET /{act_id}/adspixels. " + docMeasurementWorkflow,
			InputSchema:        listDatasetsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in := parseListPagination(params, defaultDatasetListFields)
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.listDatasets(ctx, toolMetaListDatasets, userID, actID, in)
			},
		},
		{
			Name:               toolMetaGetDataset,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Fetches one Meta pixel/dataset by id.",
			Description:        "Calls GET /{dataset_id}. Use after meta_list_datasets to inspect last_fired_time and availability.",
			InputSchema:        getDatasetSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.getDataset(ctx, toolMetaGetDataset, userID, parseGetDatasetInput(params))
			},
		},
		{
			Name:               toolMetaListCreativeAds,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists ads that use a given creative.",
			Description:        "Calls GET /{creative_id}/ads. Creative governance: meta_list_creatives → meta_list_creative_ads.",
			InputSchema:        listCreativeAdsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseListCreativeAdsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.listCreativeAds(ctx, toolMetaListCreativeAds, userID, in)
			},
		},
		{
			Name:               toolMetaGetAccountActivities,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Returns account-level change history for a Meta ad account.",
			Description:        "Calls GET /{act_id}/activities. " + docActivitiesTime,
			InputSchema:        accountActivitiesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseActivitiesInput(params)
				if err != nil {
					return nil, err
				}
				actID := strings.TrimSpace(toString(params["act_id"]))
				return svc.getAccountActivities(ctx, toolMetaGetAccountActivities, userID, actID, in)
			},
		},
		{
			Name:               toolMetaSearchBehaviors,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Meta behavior targeting categories.",
			Description:        "Calls GET /search?type=adTargetingCategory&class=behaviors. Targeting research after interests/geo.",
			InputSchema:        searchBehaviorsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				return svc.searchBehaviors(ctx, toolMetaSearchBehaviors, userID, parseSearchLimit(params))
			},
		},
		{
			Name:               toolMetaSearchDemographics,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Meta demographic targeting categories.",
			Description:        "Calls GET /search?type=adTargetingCategory&class={class}. " + docDemographicClass,
			InputSchema:        searchDemographicsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				class := parseDemographicClass(params)
				return svc.searchDemographics(ctx, toolMetaSearchDemographics, userID, class, parseSearchLimit(params))
			},
		},
		{
			Name:               toolMetaGetInterestSuggestions,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Suggests related interests from seed interest names.",
			Description:        "Calls GET /search?type=adinterestsuggestion. " + docInterestSuggestions,
			InputSchema:        interestSuggestionsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseInterestSuggestionsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getInterestSuggestions(ctx, toolMetaGetInterestSuggestions, userID, in)
			},
		},
		{
			Name:               toolMetaGetDatasetStats,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Returns event firing stats for a Meta pixel/dataset.",
			Description:        "Calls GET /{dataset_id}/stats. Signal health check — is Purchase (or another event) firing?",
			InputSchema:        datasetStatsSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseDatasetStatsInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getDatasetStats(ctx, toolMetaGetDatasetStats, userID, in)
			},
		},
		{
			Name:               toolMetaGetDatasetQuality,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Returns Dataset Quality API metrics (e.g. EMQ) for a pixel.",
			Description:        "Calls GET /dataset_quality?dataset_id=.... " + docDatasetQualityNote,
			InputSchema:        datasetQualitySchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseDatasetQualityInput(params)
				if err != nil {
					return nil, err
				}
				return svc.getDatasetQuality(ctx, toolMetaGetDatasetQuality, userID, in)
			},
		},
		{
			Name:               toolMetaGetAdSetActivities,
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Returns change history for one Meta ad set.",
			Description:        "Calls GET /{adset_id}/activities. " + docActivitiesTime,
			InputSchema:        adSetActivitiesSchema(),
			RequiresConnection: true,
			Execute: func(ctx context.Context, userID string, params map[string]any) (any, error) {
				in, err := parseActivitiesInput(params)
				if err != nil {
					return nil, err
				}
				adSetID := strings.TrimSpace(toString(params["adset_id"]))
				return svc.getAdSetActivities(ctx, toolMetaGetAdSetActivities, userID, adSetID, in)
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
