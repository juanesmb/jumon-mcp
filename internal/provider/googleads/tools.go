package googleads

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	"jumon-mcp/internal/provider/registry"
)

const platformName = "google"

type Config struct {
	APIVersion string
}

func RegisterTools(reg *registry.Registry, gatewayClient *gateway.Client, config Config) error {
	if strings.TrimSpace(config.APIVersion) == "" {
		config.APIVersion = "v22"
	}

	tools := []registry.ToolDefinition{
		{
			Name:               "google_list_ad_accounts",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Google Ads customer IDs accessible to the connected user.",
			Description:        "Calls customers:listAccessibleCustomers and returns accessible customer resource names.",
			InputSchema:        map[string]any{"type": "object"},
			RequiresConnection: true,
			Execute:            listAdAccountsExecutor(gatewayClient, config),
		},
		{
			Name:               "google_list_client_accounts_under_manager",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists non-manager client accounts under one manager account.",
			Description:        "Runs a GAQL customer_client query under an MCC manager account.",
			InputSchema:        listClientAccountsSchema(),
			RequiresConnection: true,
			Execute:            listClientAccountsExecutor(gatewayClient, config),
		},
		{
			Name:               "google_search_campaigns",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Google Ads campaigns and core metrics.",
			Description:        "Runs a GAQL campaign query with optional status/name/date filters.",
			InputSchema:        searchCampaignsSchema(),
			RequiresConnection: true,
			Execute:            searchCampaignsExecutor(gatewayClient, config),
		},
		{
			Name:               "google_search_ad_groups",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Google Ads ad groups and key metrics.",
			Description:        "Runs a GAQL ad_group query with optional campaign/status/date filters.",
			InputSchema:        searchAdGroupsSchema(),
			RequiresConnection: true,
			Execute:            searchAdGroupsExecutor(gatewayClient, config),
		},
		{
			Name:               "google_search_ads",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Searches Google Ads ad-level entities and metrics.",
			Description:        "Runs a GAQL ad_group_ad query with optional campaign/adgroup/status/date filters.",
			InputSchema:        searchAdsSchema(),
			RequiresConnection: true,
			Execute:            searchAdsExecutor(gatewayClient, config),
		},
	}

	for _, tool := range tools {
		if err := reg.Register(tool); err != nil {
			return err
		}
	}
	return nil
}

func listAdAccountsExecutor(gatewayClient *gateway.Client, config Config) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		_ = params
		path := fmt.Sprintf("%s/customers:listAccessibleCustomers", config.APIVersion)
		return proxyGoogleJSON(ctx, gatewayClient, userID, "GET", path, nil, nil)
	}
}

func listClientAccountsExecutor(gatewayClient *gateway.Client, config Config) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		customerID := normalizeCustomerID(params["customer_id"])
		if customerID == "" {
			return nil, fmt.Errorf("customer_id is required")
		}
		query := strings.Join([]string{
			"SELECT customer_client.id, customer_client.descriptive_name, customer_client.currency_code, customer_client.time_zone, customer_client.manager",
			"FROM customer_client",
			"WHERE customer_client.manager = false",
			"ORDER BY customer_client.id",
		}, " ")
		loginID := normalizeCustomerID(params["login_customer_id"])
		return googleSearch(ctx, gatewayClient, config.APIVersion, userID, customerID, loginID, query)
	}
}

func searchCampaignsExecutor(gatewayClient *gateway.Client, config Config) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		customerID := normalizeCustomerID(params["customer_id"])
		if customerID == "" {
			return nil, fmt.Errorf("customer_id is required")
		}
		query := strings.Join([]string{
			"SELECT campaign.id, campaign.name, campaign.status, campaign.advertising_channel_type, campaign_budget.amount_micros, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions",
			"FROM campaign",
			buildWhereClause([]string{
				inClause("campaign.id", toStringSlice(params["campaign_ids"])),
				likeClause("campaign.name", toString(params["campaign_name_contains"])),
				enumInClause("campaign.status", toStringSlice(params["statuses"])),
				dateBetweenClause(toString(params["date_range_start"]), toString(params["date_range_end"])),
			}),
			"ORDER BY campaign.id DESC",
		}, " ")
		loginID := normalizeCustomerID(params["login_customer_id"])
		return googleSearch(ctx, gatewayClient, config.APIVersion, userID, customerID, loginID, query)
	}
}

func searchAdGroupsExecutor(gatewayClient *gateway.Client, config Config) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		customerID := normalizeCustomerID(params["customer_id"])
		if customerID == "" {
			return nil, fmt.Errorf("customer_id is required")
		}
		query := strings.Join([]string{
			"SELECT ad_group.id, ad_group.name, ad_group.status, ad_group.campaign, campaign.name, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions",
			"FROM ad_group",
			buildWhereClause([]string{
				inClause("ad_group.id", toStringSlice(params["ad_group_ids"])),
				inClause("campaign.id", toStringSlice(params["campaign_ids"])),
				enumInClause("ad_group.status", toStringSlice(params["statuses"])),
				dateBetweenClause(toString(params["date_range_start"]), toString(params["date_range_end"])),
			}),
			"ORDER BY ad_group.id DESC",
		}, " ")
		loginID := normalizeCustomerID(params["login_customer_id"])
		return googleSearch(ctx, gatewayClient, config.APIVersion, userID, customerID, loginID, query)
	}
}

func searchAdsExecutor(gatewayClient *gateway.Client, config Config) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		customerID := normalizeCustomerID(params["customer_id"])
		if customerID == "" {
			return nil, fmt.Errorf("customer_id is required")
		}
		query := strings.Join([]string{
			"SELECT ad_group_ad.ad.id, ad_group_ad.status, ad_group_ad.ad.type, ad_group.id, campaign.id, campaign.name, metrics.clicks, metrics.impressions, metrics.cost_micros, metrics.conversions",
			"FROM ad_group_ad",
			buildWhereClause([]string{
				inClause("ad_group_ad.ad.id", toStringSlice(params["ad_ids"])),
				inClause("ad_group.id", toStringSlice(params["ad_group_ids"])),
				inClause("campaign.id", toStringSlice(params["campaign_ids"])),
				enumInClause("ad_group_ad.status", toStringSlice(params["statuses"])),
				dateBetweenClause(toString(params["date_range_start"]), toString(params["date_range_end"])),
			}),
			"ORDER BY ad_group_ad.ad.id DESC",
		}, " ")
		loginID := normalizeCustomerID(params["login_customer_id"])
		return googleSearch(ctx, gatewayClient, config.APIVersion, userID, customerID, loginID, query)
	}
}

func googleSearch(ctx context.Context, gatewayClient *gateway.Client, version, userID, customerID, loginCustomerID, query string) (any, error) {
	path := fmt.Sprintf("%s/customers/%s/googleAds:search", version, customerID)
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
	return proxyGoogleJSON(ctx, gatewayClient, userID, "POST", path, body, headers)
}

func proxyGoogleJSON(ctx context.Context, gatewayClient *gateway.Client, userID, method, path string, body any, headers map[string]string) (any, error) {
	resp, err := gatewayClient.ProxyProviderOrRefresh(ctx, platformName, userID, method, path, nil, body, headers)
	if err != nil {
		return nil, err
	}
	if gateway.IsNotConnectedResponse(resp) {
		return nil, &catalog.PlatformNotConnectedError{Platform: platformName, ConnectURL: gatewayClient.ConnectURLHint()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(resp.Body)))
	}
	var payload any
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return map[string]any{
			"status": resp.StatusCode,
			"body":   string(resp.Body),
		}, nil
	}
	return payload, nil
}

func normalizeCustomerID(raw any) string {
	value := strings.TrimSpace(toString(raw))
	return strings.TrimPrefix(value, "customers/")
}

func buildWhereClause(parts []string) string {
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

func inClause(field string, values []string) string {
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

func enumInClause(field string, values []string) string {
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

func likeClause(field, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	escaped := strings.ReplaceAll(trimmed, "'", "\\'")
	return fmt.Sprintf("%s LIKE '%%%s%%'", field, escaped)
}

func dateBetweenClause(start, end string) string {
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

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func toStringSlice(value any) []string {
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

func listClientAccountsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{
			"customer_id",
		},
		"properties": map[string]any{
			"customer_id":       map[string]any{"type": "string"},
			"login_customer_id": map[string]any{"type": "string"},
		},
	}
}

func searchCampaignsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{
			"customer_id",
		},
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
		"type": "object",
		"required": []string{
			"customer_id",
		},
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
		"type": "object",
		"required": []string{
			"customer_id",
		},
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
