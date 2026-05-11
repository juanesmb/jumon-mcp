package linkedin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
)

type linkedinUpstreamPort interface {
	requestJSON(
		ctx context.Context,
		userID, mcpTool, method, path string,
		query map[string]string,
		body any,
		extraHeaders map[string]string,
	) (any, error)
}

type linkedinGateway struct {
	client *gateway.Client
}

func newLinkedInGateway(client *gateway.Client) linkedinUpstreamPort {
	return &linkedinGateway{client: client}
}

func (g *linkedinGateway) requestJSON(
	ctx context.Context,
	userID, mcpTool, method, path string,
	query map[string]string,
	body any,
	extraHeaders map[string]string,
) (any, error) {
	headers := map[string]string{
		"Linkedin-Version":           "202504",
		"X-Restli-Protocol-Version": "2.0.0",
	}
	for key, value := range extraHeaders {
		headers[key] = value
	}

	resp, err := g.client.ProxyProviderOrRefresh(ctx, platformName, mcpTool, userID, method, path, query, body, headers)
	if err != nil {
		return nil, err
	}

	if gateway.IsNotConnectedResponse(resp) {
		return nil, &catalog.PlatformNotConnectedError{Platform: platformName, ConnectURL: g.client.ConnectURLHint()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("linkedin api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(resp.Body)))
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

