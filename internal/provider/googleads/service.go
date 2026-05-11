package googleads

import (
	"context"
	"fmt"
	"strings"

	"jumon-mcp/internal/infrastructure/gateway"
)

type service struct {
	proxy      googleUpstreamPort
	apiVersion string
}

func newGoogleService(client *gateway.Client, config Config) *service {
	return &service{
		proxy:      newGoogleGateway(client),
		apiVersion: config.APIVersion,
	}
}

type listClientAccountsInput struct {
	customerID       string
	loginCustomerID  string
}

type searchCampaignsInput struct {
	customerID          string
	loginCustomerID     string
	campaignIDs         []string
	campaignNameContains string
	statuses            []string
	dateRangeStart     string
	dateRangeEnd       string
}

type searchAdGroupsInput struct {
	customerID      string
	loginCustomerID string
	adGroupIDs      []string
	campaignIDs     []string
	statuses        []string
	dateRangeStart string
	dateRangeEnd   string
}

type searchAdsInput struct {
	customerID      string
	loginCustomerID string
	adIDs           []string
	adGroupIDs      []string
	campaignIDs     []string
	statuses        []string
	dateRangeStart string
	dateRangeEnd   string
}

func (s *service) listAdAccounts(ctx context.Context, userID, mcpTool string) (any, error) {
	path := pathListAccessibleCustomers(s.apiVersion)
	return s.proxy.requestJSON(ctx, userID, mcpTool, "GET", path, nil, nil)
}

func (s *service) listClientAccounts(ctx context.Context, userID, mcpTool string, in listClientAccountsInput) (any, error) {
	query := strings.Join([]string{
		"SELECT customer_client.id, customer_client.descriptive_name, customer_client.currency_code, customer_client.time_zone, customer_client.manager",
		"FROM customer_client",
		"WHERE customer_client.manager = false",
		"ORDER BY customer_client.id",
	}, " ")
	return s.googleSearch(ctx, userID, mcpTool, in.customerID, in.loginCustomerID, query)
}

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

func (s *service) googleSearch(
	ctx context.Context,
	userID, mcpTool,
	customerID, loginCustomerID,
	query string,
) (any, error) {
	path := pathGoogleAdsSearch(s.apiVersion, customerID)
	headers := map[string]string{}
	if loginCustomerID != "" {
		headers["login-customer-id"] = loginCustomerID
	}
	body := map[string]any{
		"query": query,
		"searchSettings": map[string]any{
			"returnTotalResultsCount": true,
		},
	}
	return s.proxy.requestJSON(ctx, userID, mcpTool, "POST", path, body, headers)
}

func parseListClientAccountsInput(params map[string]any) (listClientAccountsInput, error) {
	customerID := googleNormalizeCustomerID(params["customer_id"])
	if customerID == "" {
		return listClientAccountsInput{}, fmt.Errorf("customer_id is required")
	}
	loginID := googleNormalizeCustomerID(params["login_customer_id"])
	return listClientAccountsInput{
		customerID:      customerID,
		loginCustomerID: loginID,
	}, nil
}

func parseSearchCampaignsInput(params map[string]any) (searchCampaignsInput, error) {
	customerID := googleNormalizeCustomerID(params["customer_id"])
	if customerID == "" {
		return searchCampaignsInput{}, fmt.Errorf("customer_id is required")
	}

	return searchCampaignsInput{
		customerID:           customerID,
		loginCustomerID:      googleNormalizeCustomerID(params["login_customer_id"]),
		campaignIDs:          googleToStringSlice(params["campaign_ids"]),
		campaignNameContains: googleToString(params["campaign_name_contains"]),
		statuses:             googleToStringSlice(params["statuses"]),
		dateRangeStart:       googleToString(params["date_range_start"]),
		dateRangeEnd:         googleToString(params["date_range_end"]),
	}, nil
}

func parseSearchAdGroupsInput(params map[string]any) (searchAdGroupsInput, error) {
	customerID := googleNormalizeCustomerID(params["customer_id"])
	if customerID == "" {
		return searchAdGroupsInput{}, fmt.Errorf("customer_id is required")
	}
	return searchAdGroupsInput{
		customerID:      customerID,
		loginCustomerID: googleNormalizeCustomerID(params["login_customer_id"]),
		adGroupIDs:      googleToStringSlice(params["ad_group_ids"]),
		campaignIDs:     googleToStringSlice(params["campaign_ids"]),
		statuses:        googleToStringSlice(params["statuses"]),
		dateRangeStart: googleToString(params["date_range_start"]),
		dateRangeEnd:   googleToString(params["date_range_end"]),
	}, nil
}

func parseSearchAdsInput(params map[string]any) (searchAdsInput, error) {
	customerID := googleNormalizeCustomerID(params["customer_id"])
	if customerID == "" {
		return searchAdsInput{}, fmt.Errorf("customer_id is required")
	}
	return searchAdsInput{
		customerID:      customerID,
		loginCustomerID: googleNormalizeCustomerID(params["login_customer_id"]),
		adIDs:           googleToStringSlice(params["ad_ids"]),
		adGroupIDs:      googleToStringSlice(params["ad_group_ids"]),
		campaignIDs:     googleToStringSlice(params["campaign_ids"]),
		statuses:        googleToStringSlice(params["statuses"]),
		dateRangeStart: googleToString(params["date_range_start"]),
		dateRangeEnd:   googleToString(params["date_range_end"]),
	}, nil
}

func googleNormalizeCustomerID(raw any) string {
	value := strings.TrimSpace(googleToString(raw))
	return strings.TrimPrefix(value, "customers/")
}

func googleBuildWhereClause(parts []string) string {
	filters := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			filters = append(filters, trimmed)
		}
	}
	if len(filters) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(filters, " AND ")
}

func googleInClause(field string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	clean := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	if len(clean) == 0 {
		return ""
	}
	return fmt.Sprintf("%s IN (%s)", field, strings.Join(clean, ","))
}

func googleEnumInClause(field string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	clean := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized != "" {
			clean = append(clean, normalized)
		}
	}
	if len(clean) == 0 {
		return ""
	}
	return fmt.Sprintf("%s IN (%s)", field, strings.Join(clean, ","))
}

func googleLikeClause(field, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	escaped := strings.ReplaceAll(trimmed, "'", "\\'")
	return fmt.Sprintf("%s LIKE '%%%s%%'", field, escaped)
}

func googleDateBetweenClause(start, end string) string {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" && end == "" {
		return ""
	}
	if start == "" {
		return fmt.Sprintf("segments.date <= '%s'", end)
	}
	if end == "" {
		return fmt.Sprintf("segments.date >= '%s'", start)
	}
	return fmt.Sprintf("segments.date BETWEEN '%s' AND '%s'", start, end)
}

func googleToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func googleToStringSlice(value any) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{v}
	default:
		return nil
	}
}

