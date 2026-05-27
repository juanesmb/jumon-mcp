package googleads

import (
	"context"
	"strings"
)

func (s *service) searchCampaigns(ctx context.Context, userID, mcpTool string, in searchCampaignsInput) (any, error) {
	query := strings.Join([]string{
		"SELECT campaign.id, campaign.name, campaign.status, campaign.advertising_channel_type, campaign_budget.amount_micros, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions",
		"FROM campaign",
		googleBuildWhereClause([]string{
			googleInClause("campaign.id", in.campaignIDs),
			googleLikeClause("campaign.name", in.campaignNameContains),
			googleEnumInClause("campaign.status", in.statuses),
			googleDateBetweenClause(in.dateRangeStart, in.dateRangeEnd),
		}),
		"ORDER BY campaign.id DESC",
	}, " ")
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, query)
}

func (s *service) searchAdGroups(ctx context.Context, userID, mcpTool string, in searchAdGroupsInput) (any, error) {
	query := strings.Join([]string{
		"SELECT ad_group.id, ad_group.name, ad_group.status, ad_group.campaign, campaign.name, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions",
		"FROM ad_group",
		googleBuildWhereClause([]string{
			googleInClause("ad_group.id", in.adGroupIDs),
			googleInClause("campaign.id", in.campaignIDs),
			googleEnumInClause("ad_group.status", in.statuses),
			googleDateBetweenClause(in.dateRangeStart, in.dateRangeEnd),
		}),
		"ORDER BY ad_group.id DESC",
	}, " ")
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, query)
}

func (s *service) searchAds(ctx context.Context, userID, mcpTool string, in searchAdsInput) (any, error) {
	query := strings.Join([]string{
		"SELECT ad_group_ad.ad.id, ad_group_ad.status, ad_group_ad.ad.type, ad_group.id, campaign.id, campaign.name, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions",
		"FROM ad_group_ad",
		googleBuildWhereClause([]string{
			googleInClause("ad_group_ad.ad.id", in.adIDs),
			googleInClause("ad_group.id", in.adGroupIDs),
			googleInClause("campaign.id", in.campaignIDs),
			googleEnumInClause("ad_group_ad.status", in.statuses),
			googleDateBetweenClause(in.dateRangeStart, in.dateRangeEnd),
		}),
		"ORDER BY ad_group_ad.ad.id DESC",
	}, " ")
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, query)
}
