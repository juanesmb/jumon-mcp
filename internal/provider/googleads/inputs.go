package googleads

import (
	"fmt"
	"strings"
	"time"
)

const (
	maxAccessibleAccounts = 50
	maxManagerScan        = 5
)

type customerContext struct {
	customerID      string
	loginCustomerID string
}

type listClientAccountsInput struct {
	customerContext
	clientNameContains string
}

type resolveCustomerInput struct {
	accountName         string
	matchMode           string
	searchUnderManagers bool
}

type searchCampaignsInput struct {
	customerContext
	campaignIDs          []string
	campaignNameContains string
	statuses             []string
	dateRangeStart       string
	dateRangeEnd         string
}

type searchAdGroupsInput struct {
	customerContext
	adGroupIDs     []string
	campaignIDs    []string
	statuses       []string
	dateRangeStart string
	dateRangeEnd   string
}

type searchAdsInput struct {
	customerContext
	adIDs          []string
	adGroupIDs     []string
	campaignIDs    []string
	statuses       []string
	dateRangeStart string
	dateRangeEnd   string
}

type reportFilters struct {
	customerContext
	campaignIDs           []string
	adGroupIDs            []string
	statuses              []string
	dateRangeStart        string
	dateRangeEnd          string
	limit                 int
	keywordContains       string
	searchTermContains    string
	nameContains          string
	conversionActionIDs   []string
}

type accountRecord struct {
	CustomerID      string  `json:"customer_id"`
	DescriptiveName string  `json:"descriptive_name"`
	Manager         bool    `json:"manager"`
	CurrencyCode    string  `json:"currency_code,omitempty"`
	TimeZone        string  `json:"time_zone,omitempty"`
	LoginCustomerID *string `json:"login_customer_id"`
}

type customerMatch struct {
	CustomerID      string `json:"customer_id"`
	DescriptiveName string `json:"descriptive_name"`
	Manager         bool   `json:"manager"`
	LoginCustomerID string `json:"login_customer_id,omitempty"`
	MatchType       string `json:"match_type"`
}

func parseCustomerContext(params map[string]any) (customerContext, error) {
	customerID := googleNormalizeCustomerID(params["customer_id"])
	if customerID == "" {
		return customerContext{}, fmt.Errorf("customer_id is required")
	}
	return customerContext{
		customerID:      customerID,
		loginCustomerID: googleNormalizeCustomerID(params["login_customer_id"]),
	}, nil
}

func parseListClientAccountsInput(params map[string]any) (listClientAccountsInput, error) {
	ctx, err := parseCustomerContext(params)
	if err != nil {
		return listClientAccountsInput{}, err
	}
	return listClientAccountsInput{
		customerContext:    ctx,
		clientNameContains: googleToString(params["client_name_contains"]),
	}, nil
}

func parseResolveCustomerInput(params map[string]any) (resolveCustomerInput, error) {
	name := strings.TrimSpace(googleToString(params["account_name"]))
	if name == "" {
		return resolveCustomerInput{}, fmt.Errorf("account_name is required")
	}
	mode := strings.ToLower(strings.TrimSpace(googleToString(params["match_mode"])))
	if mode == "" {
		mode = "contains"
	}
	if mode != "contains" && mode != "exact" {
		return resolveCustomerInput{}, fmt.Errorf("match_mode must be contains or exact")
	}
	searchUnderManagers := true
	if raw, ok := params["search_under_managers"]; ok {
		switch v := raw.(type) {
		case bool:
			searchUnderManagers = v
		}
	}
	return resolveCustomerInput{
		accountName:         name,
		matchMode:           mode,
		searchUnderManagers: searchUnderManagers,
	}, nil
}

func parseSearchCampaignsInput(params map[string]any) (searchCampaignsInput, error) {
	ctx, err := parseCustomerContext(params)
	if err != nil {
		return searchCampaignsInput{}, err
	}
	return searchCampaignsInput{
		customerContext:      ctx,
		campaignIDs:          googleToStringSlice(params["campaign_ids"]),
		campaignNameContains: googleToString(params["campaign_name_contains"]),
		statuses:             googleToStringSlice(params["statuses"]),
		dateRangeStart:       googleToString(params["date_range_start"]),
		dateRangeEnd:         googleToString(params["date_range_end"]),
	}, nil
}

func parseSearchAdGroupsInput(params map[string]any) (searchAdGroupsInput, error) {
	ctx, err := parseCustomerContext(params)
	if err != nil {
		return searchAdGroupsInput{}, err
	}
	return searchAdGroupsInput{
		customerContext: ctx,
		adGroupIDs:      googleToStringSlice(params["ad_group_ids"]),
		campaignIDs:     googleToStringSlice(params["campaign_ids"]),
		statuses:        googleToStringSlice(params["statuses"]),
		dateRangeStart:  googleToString(params["date_range_start"]),
		dateRangeEnd:    googleToString(params["date_range_end"]),
	}, nil
}

func parseSearchAdsInput(params map[string]any) (searchAdsInput, error) {
	ctx, err := parseCustomerContext(params)
	if err != nil {
		return searchAdsInput{}, err
	}
	return searchAdsInput{
		customerContext: ctx,
		adIDs:           googleToStringSlice(params["ad_ids"]),
		adGroupIDs:      googleToStringSlice(params["ad_group_ids"]),
		campaignIDs:     googleToStringSlice(params["campaign_ids"]),
		statuses:        googleToStringSlice(params["statuses"]),
		dateRangeStart:  googleToString(params["date_range_start"]),
		dateRangeEnd:    googleToString(params["date_range_end"]),
	}, nil
}

func parseReportFilters(params map[string]any) (reportFilters, error) {
	ctx, err := parseCustomerContext(params)
	if err != nil {
		return reportFilters{}, err
	}
	return reportFilters{
		customerContext:     ctx,
		campaignIDs:         googleToStringSlice(params["campaign_ids"]),
		adGroupIDs:          googleToStringSlice(params["ad_group_ids"]),
		statuses:            googleToStringSlice(params["statuses"]),
		dateRangeStart:      googleToString(params["date_range_start"]),
		dateRangeEnd:        googleToString(params["date_range_end"]),
		limit:               parseLimitParam(params),
		keywordContains:     googleToString(params["keyword_contains"]),
		searchTermContains:  googleToString(params["search_term_contains"]),
		nameContains:        googleToString(params["name_contains"]),
		conversionActionIDs: googleToStringSlice(params["conversion_action_ids"]),
	}, nil
}

func parseConversionPerformanceFilters(params map[string]any) (reportFilters, error) {
	filters, err := parseReportFilters(params)
	if err != nil {
		return reportFilters{}, err
	}
	if filters.dateRangeStart == "" && filters.dateRangeEnd == "" {
		end := time.Now().UTC()
		start := end.AddDate(0, 0, -30)
		filters.dateRangeStart = start.Format("2006-01-02")
		filters.dateRangeEnd = end.Format("2006-01-02")
	}
	return filters, nil
}

func googleNormalizeCustomerID(raw any) string {
	value := strings.TrimSpace(googleToString(raw))
	value = strings.TrimPrefix(value, "customers/")
	value = strings.ReplaceAll(value, "-", "")
	return value
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

func matchAccountName(descriptiveName, query, mode string) bool {
	name := strings.TrimSpace(descriptiveName)
	q := strings.TrimSpace(query)
	if name == "" || q == "" {
		return false
	}
	nameLower := strings.ToLower(name)
	qLower := strings.ToLower(q)
	switch mode {
	case "exact":
		return nameLower == qLower
	default:
		return strings.Contains(nameLower, qLower)
	}
}
