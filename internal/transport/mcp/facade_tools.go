package mcptransport

import (
	"context"
	"encoding/json"
	"fmt"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/middleware"
	catalogusecase "jumon-mcp/internal/usecase/catalog"
	executionusecase "jumon-mcp/internal/usecase/execution"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FacadeTools struct {
	catalogService   *catalogusecase.Service
	executionService *executionusecase.Service
}

func NewFacadeTools(catalogService *catalogusecase.Service, executionService *executionusecase.Service) *FacadeTools {
	return &FacadeTools{
		catalogService:   catalogService,
		executionService: executionService,
	}
}

type ExplorePlatformInput struct {
	Platform  string   `json:"platform,omitempty"`
	ToolNames []string `json:"tool_names,omitempty"`
}

type ExecutePlatformToolInput struct {
	ToolName       string         `json:"tool_name"`
	ToolParameters map[string]any `json:"tool_parameters,omitempty"`
}

func (f *FacadeTools) ExplorePlatform(ctx context.Context, req *mcp.CallToolRequest, input ExplorePlatformInput) (*mcp.CallToolResult, catalogusecase.ExploreOutput, error) {
	_ = req
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return &mcp.CallToolResult{}, catalogusecase.ExploreOutput{}, fmt.Errorf("missing authenticated user in request context")
	}

	output, err := f.catalogService.Explore(ctx, userID, catalogusecase.ExploreInput{
		Platform:  input.Platform,
		ToolNames: input.ToolNames,
	})
	if err != nil {
		return &mcp.CallToolResult{}, catalogusecase.ExploreOutput{}, err
	}
	return marshalToolOutput(output)
}

func (f *FacadeTools) ExecutePlatformTool(ctx context.Context, req *mcp.CallToolRequest, input ExecutePlatformToolInput) (*mcp.CallToolResult, map[string]any, error) {
	_ = req
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return &mcp.CallToolResult{}, nil, fmt.Errorf("missing authenticated user in request context")
	}

	output, err := f.executionService.Execute(ctx, userID, executionusecase.ExecuteInput{
		ToolName:       input.ToolName,
		ToolParameters: input.ToolParameters,
	})
	if err != nil {
		if notConnected, ok := err.(*catalog.PlatformNotConnectedError); ok {
			result := map[string]any{
				"error":       "platform_not_connected",
				"platform":    notConnected.Platform,
				"message":     fmt.Sprintf("%s is not connected for this user.", notConnected.Platform),
				"connect_url": notConnected.ConnectURL,
			}
			toolResult, _, marshalErr := marshalToolOutput(result)
			return toolResult, result, marshalErr
		}
		return &mcp.CallToolResult{}, nil, err
	}

	result := map[string]any{
		"tool_name": output.Result.ToolName,
		"data":      output.Result.Data,
	}
	toolResult, _, marshalErr := marshalToolOutput(result)
	return toolResult, result, marshalErr
}

func marshalToolOutput[T any](payload T) (*mcp.CallToolResult, T, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		var zero T
		return &mcp.CallToolResult{}, zero, fmt.Errorf("marshal tool output: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(encoded)},
		},
	}, payload, nil
}
