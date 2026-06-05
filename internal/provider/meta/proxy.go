package meta

import (
	"context"
	"encoding/json"
	"fmt"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	infrahttp "jumon-mcp/internal/infrastructure/http"
)

type metaUpstreamPort interface {
	getWithRefresh(ctx context.Context, mcpTool, userID, path string, query map[string]string) (json.RawMessage, error)
	postFormWithRefresh(ctx context.Context, mcpTool, userID, path string, query map[string]string, body map[string]any) (json.RawMessage, error)
}

type metaGateway struct {
	client *gateway.Client
}

func newMetaGateway(c *gateway.Client) metaUpstreamPort {
	return &metaGateway{client: c}
}

func (m *metaGateway) getWithRefresh(ctx context.Context, mcpTool, userID, path string, query map[string]string) (json.RawMessage, error) {
	resp, err := m.client.ProxyProviderOrRefresh(ctx, platformName, mcpTool, userID, "GET", path, query, nil, nil)
	return decodeMetaProxy(resp, err, m.client)
}

func (m *metaGateway) postFormWithRefresh(ctx context.Context, mcpTool, userID, path string, query map[string]string, body map[string]any) (json.RawMessage, error) {
	resp, err := m.client.ProxyProviderOrRefresh(ctx, platformName, mcpTool, userID, "POST", path, query, body, nil)
	return decodeMetaProxy(resp, err, m.client)
}

func decodeMetaProxy(resp *infrahttp.Response, err error, gw *gateway.Client) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}
	if gw == nil {
		return nil, fmt.Errorf("meta: gateway client is nil")
	}
	if gateway.IsNotConnectedResponse(resp) {
		return nil, &catalog.PlatformNotConnectedError{Platform: platformName, ConnectURL: gw.ConnectURLHint()}
	}
	if resp == nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := 0
		body := ""
		if resp != nil {
			code = resp.StatusCode
			body = string(resp.Body)
		}
		return nil, formatMetaAPIError(code, body)
	}
	return append(json.RawMessage(nil), resp.Body...), nil
}
