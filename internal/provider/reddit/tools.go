package reddit

import (
	"context"
	"encoding/json"
	"strings"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	"jumon-mcp/internal/provider/registry"
)

func RegisterTools(reg *registry.Registry, gatewayClient *gateway.Client) error {
	port := newRedditGateway(gatewayClient)
	svc := newService(port)

	tools := []registry.ToolDefinition{
		{
			Name:               "reddit_list_businesses",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Reddit Advertising businesses visible to the connected user.",
			Description:        "Calls GET me/businesses (list_my_businesses). Use pagination fields in the response, then pass a business id to reddit_list_ad_accounts.",
			InputSchema:        listBusinessesSchema(),
			RequiresConnection: true,
			Execute:            listBusinessesExecutor(svc),
		},
		{
			Name:               "reddit_list_ad_accounts",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Reddit Ads accounts under one business.",
			Description:        "Calls GET businesses/{business_id}/ad_accounts after you obtain business_id from reddit_list_businesses. Supports Reddit pagination via page_token.",
			InputSchema:        listAdAccountsSchema(),
			RequiresConnection: true,
			Execute:            listAdAccountsExecutor(svc),
		},
		{
			Name:               "reddit_list_campaigns",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Reddit Ads campaigns under one ad account.",
			Description:        "Calls GET ad_accounts/{ad_account_id}/campaigns (Campaign Management Read). Requires ad_account_id from reddit_list_ad_accounts (after reddit_list_businesses). Uses Reddit pagination: page_token and mapped page.size.",
			InputSchema:        listCampaignsSchema(),
			RequiresConnection: true,
			Execute:            listCampaignsExecutor(svc),
		},
		{
			Name:               "reddit_list_ad_groups",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Reddit Ads ad groups under one ad account.",
			Description:        "Calls GET ad_accounts/{ad_account_id}/ad_groups. Requires ad_account_id from reddit_list_ad_accounts. Optional campaign_id filters by campaign. Uses Reddit pagination (page_token, page.size). Reddit id[] filters for ad groups are not passed through this tool until the gateway supports repeated query keys.",
			InputSchema:        listAdGroupsSchema(),
			RequiresConnection: true,
			Execute:            listAdGroupsExecutor(svc),
		},
		{
			Name:               "reddit_list_ads",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists Reddit Ads under one ad account.",
			Description:        "Calls GET ad_accounts/{ad_account_id}/ads. Requires ad_account_id from reddit_list_ad_accounts. Uses Reddit pagination (page_token, page.size). Reddit id[] ad filters are not passed through this tool until the gateway supports repeated query keys.",
			InputSchema:        listAdsSchema(),
			RequiresConnection: true,
			Execute:            listAdsExecutor(svc),
		},
		{
			Name:               "reddit_list_funding_instruments",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Lists funding instruments for one Reddit Ads account.",
			Description:        "Calls GET ad_accounts/{ad_account_id}/funding_instruments. Requires ad_account_id from reddit_list_ad_accounts. Optional start_time, end_time (ISO 8601), mode (ACTIVE, INACTIVE, UPCOMING, SELECTABLE, ALL), and search. Uses Reddit pagination. Billing rate limit is stricter (~30 requests per 60 seconds per Reddit docs). Array query filters (funding_instrument_ids, types) are not passed through until the gateway supports repeated query keys.",
			InputSchema:        listFundingInstrumentsSchema(),
			RequiresConnection: true,
			Execute:            listFundingInstrumentsExecutor(svc),
		},
		{
			Name:               "reddit_get_report",
			Platform:           platformName,
			Action:             catalog.ToolActionRead,
			Summary:            "Creates/fetches metrics for one Reddit Ads ad account reporting request.",
			Description:        "Calls POST ad_accounts/{ad_account_id}/reports with a JSON {\"data\":{...}} body. Requires ad_account_id from reddit_list_ad_accounts (after reddit_list_businesses). starts_at and ends_at must be hourly UTC (YYYY-MM-DDTHH:00:00Z). fields and breakdowns are Reddit metric enums (see Reddit Ads Reporting docs). Reddit reporting is quota-sensitive (~60 POSTs per rolling 60 seconds per Reddit documentation). Optionally pass page_token to fetch the next page of the report via query params page.size/page.token.",
			InputSchema:        getReportSchema(),
			RequiresConnection: true,
			Execute:            getReportExecutor(svc),
		},
	}

	for _, tool := range tools {
		if err := reg.Register(tool); err != nil {
			return err
		}
	}
	return nil
}

func listBusinessesExecutor(svc *service) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listBusinessesInput{
			adAccountID: strings.TrimSpace(toString(params["ad_account_id"])),
			role:        strings.TrimSpace(toString(params["role"])),
			pageToken:   strings.TrimSpace(toString(params["page_token"])),
		}
		if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
			in.pageSize = ps
		}

		raw, err := svc.listMyBusinesses(ctx, userID, in)
		if err != nil {
			return nil, err
		}
		return unmarshalPayload(raw)
	}
}

func listAdAccountsExecutor(svc *service) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listAdAccountsInput{
			pageToken: strings.TrimSpace(toString(params["page_token"])),
		}
		if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
			in.pageSize = ps
		}

		businessID := strings.TrimSpace(toString(params["business_id"]))
		raw, err := svc.listAdAccountsByBusiness(ctx, userID, businessID, in)
		if err != nil {
			return nil, err
		}
		return unmarshalPayload(raw)
	}
}

func listCampaignsExecutor(svc *service) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listCampaignsInput{
			pageToken: strings.TrimSpace(toString(params["page_token"])),
		}
		if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
			in.pageSize = ps
		}
		adAccountID := strings.TrimSpace(toString(params["ad_account_id"]))
		raw, err := svc.listCampaignsForAdAccount(ctx, userID, adAccountID, in)
		if err != nil {
			return nil, err
		}
		return unmarshalPayload(raw)
	}
}

func listAdGroupsExecutor(svc *service) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listAdGroupsInput{
			pageToken:  strings.TrimSpace(toString(params["page_token"])),
			campaignID: strings.TrimSpace(toString(params["campaign_id"])),
		}
		if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
			in.pageSize = ps
		}
		adAccountID := strings.TrimSpace(toString(params["ad_account_id"]))
		raw, err := svc.listAdGroupsForAdAccount(ctx, userID, adAccountID, in)
		if err != nil {
			return nil, err
		}
		return unmarshalPayload(raw)
	}
}

func listAdsExecutor(svc *service) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listAdsInput{
			pageToken: strings.TrimSpace(toString(params["page_token"])),
		}
		if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
			in.pageSize = ps
		}
		adAccountID := strings.TrimSpace(toString(params["ad_account_id"]))
		raw, err := svc.listAdsForAdAccount(ctx, userID, adAccountID, in)
		if err != nil {
			return nil, err
		}
		return unmarshalPayload(raw)
	}
}

func listFundingInstrumentsExecutor(svc *service) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := listFundingInstrumentsInput{
			pageToken: strings.TrimSpace(toString(params["page_token"])),
			startTime: strings.TrimSpace(toString(params["start_time"])),
			endTime:   strings.TrimSpace(toString(params["end_time"])),
			mode:      strings.TrimSpace(toString(params["mode"])),
			search:    strings.TrimSpace(toString(params["search"])),
		}
		if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
			in.pageSize = ps
		}
		adAccountID := strings.TrimSpace(toString(params["ad_account_id"]))
		raw, err := svc.listFundingInstrumentsForAdAccount(ctx, userID, adAccountID, in)
		if err != nil {
			return nil, err
		}
		return unmarshalPayload(raw)
	}
}

func getReportExecutor(svc *service) registry.Executor {
	return func(ctx context.Context, userID string, params map[string]any) (any, error) {
		in := createReportInput{
			pageToken:       strings.TrimSpace(toString(params["page_token"])),
			fields:          toStringSlice(params["fields"]),
			breakdowns:      toStringSlice(params["breakdowns"]),
			customColumnIDs: toStringSlice(params["custom_column_ids"]),
			filter:          strings.TrimSpace(toString(params["filter"])),
			startsAt:        strings.TrimSpace(toString(params["starts_at"])),
			endsAt:          strings.TrimSpace(toString(params["ends_at"])),
			timeZoneID:      strings.TrimSpace(toString(params["time_zone_id"])),
		}
		if ps, ok := toInt(params["page_size"]); ok && ps > 0 {
			in.pageSize = ps
		}

		adAccountID := strings.TrimSpace(toString(params["ad_account_id"]))
		raw, err := svc.createReportForAdAccount(ctx, userID, adAccountID, in)
		if err != nil {
			return nil, err
		}
		return unmarshalPayload(raw)
	}
}

func unmarshalPayload(raw json.RawMessage) (any, error) {
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func listBusinessesSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ad_account_id": map[string]any{
				"type":        "string",
				"description": "Optional. Only return businesses that grant access to this ad account.",
			},
			"role": map[string]any{
				"type":        "string",
				"description": "Optional. Filter businesses by user role (e.g. BUSINESS_ADMIN).",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Mapped to Reddit page.size (default 700, max 700 for this upstream API).",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Mapped to Reddit page.token (pagination).",
			},
		},
	}
}

func listAdAccountsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"business_id"},
		"properties": map[string]any{
			"business_id": map[string]any{
				"type":        "string",
				"description": "Business id from reddit_list_businesses (data[].id).",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Mapped to Reddit page.size (default 100, max 1000).",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Mapped to Reddit page.token (pagination).",
			},
		},
	}
}

func listCampaignsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"ad_account_id"},
		"properties": map[string]any{
			"ad_account_id": map[string]any{
				"type":        "string",
				"description": "Reddit ad account id from reddit_list_ad_accounts (typically data[].id).",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Mapped to Reddit page.size (default 100, max 1000).",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Mapped to Reddit page.token (pagination).",
			},
		},
	}
}

func listAdGroupsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"ad_account_id"},
		"properties": map[string]any{
			"ad_account_id": map[string]any{
				"type":        "string",
				"description": "Reddit ad account id from reddit_list_ad_accounts.",
			},
			"campaign_id": map[string]any{
				"type":        "string",
				"description": "Optional. Filter ad groups by campaign id.",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Mapped to Reddit page.size (default 100, max 1000).",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Mapped to Reddit page.token (pagination).",
			},
		},
	}
}

func listAdsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"ad_account_id"},
		"properties": map[string]any{
			"ad_account_id": map[string]any{
				"type":        "string",
				"description": "Reddit ad account id from reddit_list_ad_accounts.",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Mapped to Reddit page.size (default 100, max 1000).",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Mapped to Reddit page.token (pagination).",
			},
		},
	}
}

func listFundingInstrumentsSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"ad_account_id"},
		"properties": map[string]any{
			"ad_account_id": map[string]any{
				"type":        "string",
				"description": "Reddit ad account id from reddit_list_ad_accounts.",
			},
			"start_time": map[string]any{
				"type":        "string",
				"description": "Optional. Start of time window for filtering (ISO 8601), e.g. 2025-01-09T04:00:00Z",
			},
			"end_time": map[string]any{
				"type":        "string",
				"description": "Optional. End of time window for filtering (ISO 8601), e.g. 2025-02-09T04:00:00Z",
			},
			"mode": map[string]any{
				"type":        "string",
				"description": "Optional. ACTIVE, INACTIVE, UPCOMING, SELECTABLE, or ALL.",
			},
			"search": map[string]any{
				"type":        "string",
				"description": "Optional. Search text.",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Mapped to Reddit page.size (default 100, max 1000).",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Mapped to Reddit page.token (pagination).",
			},
		},
	}
}

func getReportSchema() map[string]any {
	fieldsItems := map[string]any{"type": "string"}
	return map[string]any{
		"type":     "object",
		"required": []string{"ad_account_id", "starts_at", "ends_at", "fields"},
		"properties": map[string]any{
			"ad_account_id": map[string]any{
				"type":        "string",
				"description": "Reddit ad account id from reddit_list_ad_accounts.",
			},
			"starts_at": map[string]any{
				"type":        "string",
				"description": "Inclusive report start time, hourly UTC: YYYY-MM-DDTHH:00:00Z",
			},
			"ends_at": map[string]any{
				"type":        "string",
				"description": "Inclusive report end time (hour boundaries), hourly UTC: YYYY-MM-DDTHH:00:00Z",
			},
			"fields": map[string]any{
				"type":        "array",
				"items":       fieldsItems,
				"description": "Reddit metric field names required by the Reporting API (e.g. SPEND, CLICKS); see Reddit docs for the allowed list.",
			},
			"breakdowns": map[string]any{
				"type":        "array",
				"items":       fieldsItems,
				"description": "Optional Reddit breakdown dimensions (e.g. DATE, CAMPAIGN_ID); up to limits per Reddit docs.",
			},
			"filter": map[string]any{
				"type":        "string",
				"description": "Optional comma-separated filter expression per Reddit filter-reporting-metrics documentation.",
			},
			"time_zone_id": map[string]any{
				"type":        "string",
				"description": "Optional IANA timezone id for interpreting report times when applicable.",
			},
			"custom_column_ids": map[string]any{
				"type":        "array",
				"items":       fieldsItems,
				"description": "Optional custom column IDs from Reddit Custom Columns.",
			},
			"page_size": map[string]any{
				"type":        "number",
				"description": "Mapped to Reddit query page.size for paginated reports (default 100, max 1000 per Reddit docs pattern).",
			},
			"page_token": map[string]any{
				"type":        "string",
				"description": "Mapped to Reddit page.token when fetching additional report pages.",
			},
		},
	}
}

func toString(value any) string {
	s, _ := value.(string)
	return s
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

func toStringSlice(value any) []string {
	switch v := value.(type) {
	case nil:
		return nil
	case []string:
		out := make([]string, 0, len(v))
		for _, s := range v {
			if t := strings.TrimSpace(s); t != "" {
				out = append(out, t)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(v))
		for _, x := range v {
			if s, ok := x.(string); ok {
				if t := strings.TrimSpace(s); t != "" {
					out = append(out, t)
				}
				continue
			}
			if t := strings.TrimSpace(toString(x)); t != "" {
				out = append(out, t)
			}
		}
		return out
	default:
		return nil
	}
}
