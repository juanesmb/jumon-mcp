package registry

import (
	"context"
	"fmt"

	"jumon-mcp/internal/infrastructure/gateway"
)

type GatewayConnectionReader struct {
	gatewayClient *gateway.Client
}

func NewGatewayConnectionReader(gatewayClient *gateway.Client) *GatewayConnectionReader {
	return &GatewayConnectionReader{gatewayClient: gatewayClient}
}

func (r *GatewayConnectionReader) IsConnected(ctx context.Context, provider, userID string) (bool, error) {
	resp, err := r.gatewayClient.GetConnection(ctx, provider, userID)
	if err != nil {
		return false, fmt.Errorf("connection check for %s: %w", provider, err)
	}
	return gateway.IsProviderConnected(resp), nil
}

func (r *GatewayConnectionReader) ConnectURL() string {
	return r.gatewayClient.ConnectURLHint()
}
