package googleads

import (
	"context"

	"jumon-mcp/internal/infrastructure/gateway"
)

type service struct {
	proxy                 googleUpstreamPort
	apiVersion            string
	maxAccessibleAccounts int
	maxManagerScan        int
}

func newGoogleService(client *gateway.Client, config Config) *service {
	maxAccounts := config.MaxAccessibleAccounts
	if maxAccounts <= 0 {
		maxAccounts = defaultMaxAccessibleAccounts
	}
	maxManagers := config.MaxManagerScan
	if maxManagers <= 0 {
		maxManagers = defaultMaxManagerScan
	}
	return &service{
		proxy:                 newGoogleGateway(client),
		apiVersion:            config.APIVersion,
		maxAccessibleAccounts: maxAccounts,
		maxManagerScan:        maxManagers,
	}
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
