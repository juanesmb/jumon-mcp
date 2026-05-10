package reddit

import (
	"fmt"
	"net/url"
)

const (
	pathMeBusinesses        = "me/businesses"
	queryKeyPageSize        = "page.size"
	queryKeyPageToken       = "page.token"
	queryKeyAdAccountFilter = "ad_account_id"
	queryKeyRole            = "role"
	queryKeyCampaignID      = "campaign_id"
	queryKeyStartTime       = "start_time"
	queryKeyEndTime         = "end_time"
	queryKeyMode            = "mode"
	queryKeySearch          = "search"
	queryKeyName            = "name"
)

func pathBusinessAdAccounts(businessID string) string {
	return fmt.Sprintf("businesses/%s/ad_accounts", url.PathEscape(businessID))
}

func pathAdAccountCampaigns(adAccountID string) string {
	return fmt.Sprintf("ad_accounts/%s/campaigns", url.PathEscape(adAccountID))
}

func pathAdAccountReports(adAccountID string) string {
	return fmt.Sprintf("ad_accounts/%s/reports", url.PathEscape(adAccountID))
}

func pathAdAccountAdGroups(adAccountID string) string {
	return fmt.Sprintf("ad_accounts/%s/ad_groups", url.PathEscape(adAccountID))
}

func pathAdAccountAds(adAccountID string) string {
	return fmt.Sprintf("ad_accounts/%s/ads", url.PathEscape(adAccountID))
}

func pathAdAccountFundingInstruments(adAccountID string) string {
	return fmt.Sprintf("ad_accounts/%s/funding_instruments", url.PathEscape(adAccountID))
}

func pathAdAccountPixels(adAccountID string) string {
	return fmt.Sprintf("ad_accounts/%s/pixels", url.PathEscape(adAccountID))
}

func pathAdAccountCustomAudiences(adAccountID string) string {
	return fmt.Sprintf("ad_accounts/%s/custom_audiences", url.PathEscape(adAccountID))
}

// pathForecastingBidSuggestions is POST Generate Bid Suggestion (global path, no ad_account in URL).
func pathForecastingBidSuggestions() string {
	return "forecasting/bid_suggestions"
}
