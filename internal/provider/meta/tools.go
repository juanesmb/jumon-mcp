package meta

import (
	"jumon-mcp/internal/infrastructure/gateway"
	"jumon-mcp/internal/provider/registry"
)

// RegisterTools wires the Meta provider into the MCP registry.
// P0: no tools — explore_platform lists meta after P1 registrations.
func RegisterTools(_ *registry.Registry, _ *gateway.Client) error {
	return nil
}
