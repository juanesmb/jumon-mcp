package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	infrahttp "jumon-mcp/internal/infrastructure/http"
)

const platformName = "reddit"

// redditGETPort is the outbound Reddit gateway abstraction (application depends on this, not *gateway.Client).
type redditGETPort interface {
	getWithRefresh(ctx context.Context, userID, path string, query map[string]string) (json.RawMessage, error)
	get(ctx context.Context, userID, path string, query map[string]string) (json.RawMessage, error)
}

type redditGateway struct {
	client *gateway.Client
}

func newRedditGateway(c *gateway.Client) redditGETPort {
	return &redditGateway{client: c}
}

func (r *redditGateway) getWithRefresh(ctx context.Context, userID, path string, query map[string]string) (json.RawMessage, error) {
	resp, err := r.client.ProxyProviderOrRefresh(ctx, platformName, userID, "GET", path, query, nil, nil)
	return decodeRedditProxy(resp, err, r.client)
}

func (r *redditGateway) get(ctx context.Context, userID, path string, query map[string]string) (json.RawMessage, error) {
	resp, err := r.client.ProxyProvider(ctx, platformName, userID, "GET", path, query, nil, nil)
	return decodeRedditProxy(resp, err, r.client)
}

func decodeRedditProxy(resp *infrahttp.Response, err error, gw *gateway.Client) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}
	if gw == nil {
		return nil, fmt.Errorf("reddit: gateway client is nil")
	}
	if gateway.IsNotConnectedResponse(resp) {
		return nil, &catalog.PlatformNotConnectedError{Platform: platformName, ConnectURL: gw.ConnectURLHint()}
	}
	if resp == nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := 0
		body := ""
		if resp != nil {
			code = resp.StatusCode
			body = strings.TrimSpace(string(resp.Body))
		}
		return nil, fmt.Errorf("reddit api returned status %d: %s", code, body)
	}
	return append(json.RawMessage(nil), resp.Body...), nil
}
