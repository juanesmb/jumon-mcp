package googleads

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
)

type googleUpstreamPort interface {
	requestJSON(
		ctx context.Context,
		userID, mcpTool, method, path string,
		body any,
		headers map[string]string,
	) (any, error)
}

type googleGateway struct {
	client *gateway.Client
}

func newGoogleGateway(client *gateway.Client) googleUpstreamPort {
	return &googleGateway{client: client}
}

func (g *googleGateway) requestJSON(
	ctx context.Context,
	userID, mcpTool, method, path string,
	body any,
	headers map[string]string,
) (any, error) {
	resp, err := g.client.ProxyProviderOrRefresh(ctx, platformName, mcpTool, userID, method, path, nil, body, headers)
	if err != nil {
		return nil, err
	}
	if gateway.IsNotConnectedResponse(resp) {
		return nil, &catalog.PlatformNotConnectedError{Platform: platformName, ConnectURL: g.client.ConnectURLHint()}
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

