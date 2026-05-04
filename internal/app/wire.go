package app

import (
	_ "embed"
	"log"
	stdhttp "net/http"
	"strings"

	"jumon-mcp/internal/config"
	"jumon-mcp/internal/infrastructure/gateway"
	infrahttp "jumon-mcp/internal/infrastructure/http"
	"jumon-mcp/internal/infrastructure/observability"
	"jumon-mcp/internal/provider/googleads"
	"jumon-mcp/internal/provider/linkedin"
	"jumon-mcp/internal/provider/registry"
	mcptransport "jumon-mcp/internal/transport/mcp"
	catalogusecase "jumon-mcp/internal/usecase/catalog"
	executionusecase "jumon-mcp/internal/usecase/execution"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:embed instructions/server_instructions.md
var serverInstructions string

type components struct {
	facadeTools *mcptransport.FacadeTools
}

func initComponents(cfg config.Config, rec *observability.Recorder) (*components, error) {
	otelRT := otelhttp.NewTransport(stdhttp.DefaultTransport)
	outboundTransport := rec.HTTPTransport(otelRT)

	httpClient := infrahttp.NewClient(nil, outboundTransport)
	gatewayClient := gateway.NewClient(httpClient, cfg.Gateway.BaseURL, cfg.Gateway.InternalSecret, cfg.Gateway.ConnectURL)
	connectionReader := registry.NewGatewayConnectionReader(gatewayClient)

	toolRegistry := registry.New(connectionReader)
	if err := linkedin.RegisterTools(toolRegistry, gatewayClient); err != nil {
		return nil, err
	}
	if err := googleads.RegisterTools(toolRegistry, gatewayClient, googleads.Config{APIVersion: cfg.Gateway.GoogleAPIVersion}); err != nil {
		return nil, err
	}

	instrumented := observability.NewObservableRegistry(toolRegistry, rec)

	catalogService := catalogusecase.NewService(instrumented)
	executionService := executionusecase.NewService(instrumented)
	facadeTools := mcptransport.NewFacadeTools(catalogService, executionService, rec)

	return &components{
		facadeTools: facadeTools,
	}, nil
}

func initServer(components *components) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "Jumon",
		Version: "v1.0.0",
		Title:   "Jumon MCP Facade",
	}, &mcp.ServerOptions{
		Instructions: loadServerInstructions(),
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "explore_platform",
		Description: "Discover available platforms and tools, or load schemas for specific tools by name.",
	}, components.facadeTools.ExplorePlatform)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "execute_platform_tool",
		Description: "Execute one platform tool by name with validated parameters.",
	}, components.facadeTools.ExecutePlatformTool)

	return server
}

func loadServerInstructions() string {
	trimmed := strings.TrimSpace(serverInstructions)
	if trimmed == "" {
		log.Fatal("embedded instructions file is empty")
	}
	return trimmed
}
