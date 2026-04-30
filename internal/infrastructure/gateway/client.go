package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	infrahttp "jumon-mcp/internal/infrastructure/http"
)

const headerGatewaySecret = "x-gateway-secret"

type Client struct {
	httpClient     *infrahttp.Client
	baseURL        string
	internalSecret string
	connectURLHint string
}

func NewClient(httpClient *infrahttp.Client, baseURL, internalSecret, connectURLHint string) *Client {
	return &Client{
		httpClient:     httpClient,
		baseURL:        strings.TrimRight(baseURL, "/"),
		internalSecret: internalSecret,
		connectURLHint: connectURLHint,
	}
}

func (c *Client) ConnectURLHint() string {
	return c.connectURLHint
}

func (c *Client) GetConnection(ctx context.Context, provider, userID string) (*infrahttp.Response, error) {
	path := fmt.Sprintf("%s/api/internal/connections/%s/current?userId=%s", c.baseURL, provider, url.QueryEscape(userID))
	return c.httpClient.Get(ctx, path, c.authHeaders())
}

func (c *Client) RefreshProvider(ctx context.Context, provider, userID string) (*infrahttp.Response, error) {
	path := fmt.Sprintf("%s/api/internal/providers/%s/refresh", c.baseURL, provider)
	return c.httpClient.Post(ctx, path, map[string]string{"userId": userID}, c.authHeaders())
}

func (c *Client) ProxyProvider(ctx context.Context, provider, userID, method, proxyPath string, query map[string]string, body any, headers map[string]string) (*infrahttp.Response, error) {
	path := fmt.Sprintf("%s/api/internal/providers/%s/proxy", c.baseURL, provider)
	payload := map[string]any{
		"userId": userID,
		"method": method,
		"path":   strings.TrimLeft(proxyPath, "/"),
	}
	if len(query) > 0 {
		payload["query"] = query
	}
	if body != nil {
		payload["body"] = body
	}
	if len(headers) > 0 {
		payload["headers"] = headers
	}
	return c.httpClient.Post(ctx, path, payload, c.authHeaders())
}

func (c *Client) ProxyProviderOrRefresh(ctx context.Context, provider, userID, method, proxyPath string, query map[string]string, body any, headers map[string]string) (*infrahttp.Response, error) {
	resp, err := c.ProxyProvider(ctx, provider, userID, method, proxyPath, query, body, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 401 {
		return resp, nil
	}
	if _, err := c.RefreshProvider(ctx, provider, userID); err != nil {
		return resp, nil
	}
	return c.ProxyProvider(ctx, provider, userID, method, proxyPath, query, body, headers)
}

func IsProviderConnected(resp *infrahttp.Response) bool {
	if resp == nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return false
	}
	connected, ok := payload["connected"].(bool)
	return ok && connected
}

func IsNotConnectedResponse(resp *infrahttp.Response) bool {
	if resp == nil {
		return false
	}
	if resp.StatusCode == 404 {
		return true
	}
	if len(resp.Body) == 0 {
		return false
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return false
	}
	if connected, ok := payload["connected"].(bool); ok && !connected {
		return true
	}
	if code, ok := payload["code"].(string); ok {
		normalized := strings.ToUpper(strings.TrimSpace(code))
		if strings.Contains(normalized, "CONNECTION") && strings.Contains(normalized, "NOT") {
			return true
		}
	}
	return false
}

func (c *Client) authHeaders() map[string]string {
	return map[string]string{
		headerGatewaySecret: c.internalSecret,
		"Content-Type":      "application/json",
		"Accept":            "application/json",
	}
}
