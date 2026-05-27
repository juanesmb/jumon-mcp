package googleads

import (
	"context"

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

func (s *service) listAccessibleCustomers(ctx context.Context, userID, mcpTool string) (any, error) {
	path := pathListAccessibleCustomers(s.apiVersion)
	return s.proxy.requestJSON(ctx, userID, mcpTool, "GET", path, nil, nil)
}

func (s *service) googleAdsFieldSearch(
	ctx context.Context,
	userID, mcpTool, query, pageToken string,
) (any, error) {
	path := pathGoogleAdsFieldsSearch(s.apiVersion)
	body := map[string]any{
		"query":    query,
		"pageSize": fieldServicePageSize,
	}
	if pageToken != "" {
		body["pageToken"] = pageToken
	}
	return s.proxy.requestJSON(ctx, userID, mcpTool, "POST", path, body, nil)
}
