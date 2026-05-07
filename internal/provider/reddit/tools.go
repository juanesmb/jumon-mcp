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
