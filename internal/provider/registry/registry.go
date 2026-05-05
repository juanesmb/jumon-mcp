package registry

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"jumon-mcp/internal/domain/catalog"
)

type Executor func(ctx context.Context, userID string, params map[string]any) (any, error)

type ConnectionReader interface {
	IsConnected(ctx context.Context, provider, userID string) (bool, error)
	ConnectURL() string
}

type ToolDefinition struct {
	Name               string
	Platform           string
	Action             catalog.ToolAction
	Summary            string
	Description        string
	InputSchema        map[string]any
	RequiresConnection bool
	Execute            Executor
}

type Registry struct {
	tools            map[string]ToolDefinition
	platforms        map[string]struct{}
	connectionReader ConnectionReader
}

func New(connectionReader ConnectionReader) *Registry {
	return &Registry{
		tools:            make(map[string]ToolDefinition),
		platforms:        make(map[string]struct{}),
		connectionReader: connectionReader,
	}
}

func (r *Registry) Register(definition ToolDefinition) error {
	name := strings.TrimSpace(definition.Name)
	platform := strings.TrimSpace(definition.Platform)
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if platform == "" {
		return fmt.Errorf("tool platform cannot be empty")
	}
	if definition.Execute == nil {
		return fmt.Errorf("tool %s has nil executor", name)
	}
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	definition.Name = name
	definition.Platform = platform
	r.tools[name] = definition
	r.platforms[platform] = struct{}{}
	return nil
}

func (r *Registry) ListPlatforms(ctx context.Context, userID string) ([]catalog.Platform, error) {
	platformNames := make([]string, 0, len(r.platforms))
	for platform := range r.platforms {
		platformNames = append(platformNames, platform)
	}
	sort.Strings(platformNames)

	result := make([]catalog.Platform, 0, len(platformNames))
	for _, platform := range platformNames {
		connected, err := r.connectionReader.IsConnected(ctx, platform, userID)
		if err != nil {
			return nil, err
		}
		row := catalog.Platform{
			Name:      platform,
			Connected: connected,
		}
		if !connected {
			row.ConnectURL = r.connectionReader.ConnectURL()
		}
		result = append(result, row)
	}
	return result, nil
}

func (r *Registry) ListTools(ctx context.Context, userID, platform string) ([]catalog.ToolSummary, error) {
	normalizedPlatform := strings.TrimSpace(platform)
	if normalizedPlatform == "" {
		return nil, fmt.Errorf("platform is required")
	}
	connected, err := r.connectionReader.IsConnected(ctx, normalizedPlatform, userID)
	if err != nil {
		return nil, err
	}

	tools := make([]catalog.ToolSummary, 0)
	for _, definition := range r.tools {
		if definition.Platform != normalizedPlatform {
			continue
		}
		if definition.RequiresConnection && !connected {
			continue
		}
		tools = append(tools, catalog.ToolSummary{
			Name:     definition.Name,
			Platform: definition.Platform,
			Action:   definition.Action,
			Summary:  definition.Summary,
		})
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools, nil
}

func (r *Registry) LoadSchemas(ctx context.Context, userID string, toolNames []string) ([]catalog.ToolSchema, error) {
	if len(toolNames) == 0 {
		return nil, fmt.Errorf("tool_names is required")
	}
	out := make([]catalog.ToolSchema, 0, len(toolNames))
	for _, toolName := range toolNames {
		definition, ok := r.tools[strings.TrimSpace(toolName)]
		if !ok {
			out = append(out, catalog.ToolSchema{
				Name:        strings.TrimSpace(toolName),
				Description: "Tool not found",
				InputSchema: map[string]any{"error": "tool_not_found"},
			})
			continue
		}
		if definition.RequiresConnection {
			connected, err := r.connectionReader.IsConnected(ctx, definition.Platform, userID)
			if err != nil {
				return nil, err
			}
			if !connected {
				continue
			}
		}
		out = append(out, catalog.ToolSchema{
			Name:        definition.Name,
			Description: definition.Description,
			InputSchema: definition.InputSchema,
		})
	}
	return out, nil
}

// PlatformForTool returns the canonical platform identifier for the tool, if registered.
func (r *Registry) PlatformForTool(toolName string) (string, bool) {
	definition, ok := r.tools[strings.TrimSpace(toolName)]
	if !ok {
		return "", false
	}
	return definition.Platform, true
}

func (r *Registry) Execute(ctx context.Context, userID, toolName string, params map[string]any) (catalog.ToolExecutionResult, error) {
	definition, ok := r.tools[strings.TrimSpace(toolName)]
	if !ok {
		return catalog.ToolExecutionResult{}, fmt.Errorf("tool %q not found", toolName)
	}
	if definition.RequiresConnection {
		connected, err := r.connectionReader.IsConnected(ctx, definition.Platform, userID)
		if err != nil {
			return catalog.ToolExecutionResult{}, err
		}
		if !connected {
			return catalog.ToolExecutionResult{}, &catalog.PlatformNotConnectedError{
				Platform:   definition.Platform,
				ConnectURL: r.connectionReader.ConnectURL(),
			}
		}
	}

	result, err := definition.Execute(ctx, userID, params)
	if err != nil {
		return catalog.ToolExecutionResult{}, err
	}
	return catalog.ToolExecutionResult{
		ToolName: definition.Name,
		Data:     result,
	}, nil
}
