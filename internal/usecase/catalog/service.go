package catalog

import (
	"context"
	"fmt"
	"strings"

	domain "jumon-mcp/internal/domain/catalog"
)

type Registry interface {
	ListPlatforms(ctx context.Context, userID string) ([]domain.Platform, error)
	ListTools(ctx context.Context, userID, platform string) ([]domain.ToolSummary, error)
	LoadSchemas(ctx context.Context, userID string, toolNames []string) ([]domain.ToolSchema, error)
}

type Service struct {
	registry Registry
}

func NewService(registry Registry) *Service {
	return &Service{registry: registry}
}

type ExploreInput struct {
	Platform  string
	ToolNames []string
}

type ExploreOutput struct {
	Platform      string                    `json:"platform,omitempty"`
	Platforms     []domain.Platform         `json:"platforms,omitempty"`
	ToolsCount    int                       `json:"tools_count,omitempty"`
	Tools         []domain.ToolSummary      `json:"tools,omitempty"`
	ToolsByAction map[string][]string       `json:"tools_by_action,omitempty"`
	ToolSchemas   []domain.ToolSchema       `json:"tools_with_schema,omitempty"`
	Message       string                    `json:"message,omitempty"`
	NextStep      map[string]any            `json:"next_step,omitempty"`
}

func (s *Service) Explore(ctx context.Context, userID string, input ExploreInput) (ExploreOutput, error) {
	if len(input.ToolNames) > 0 {
		schemas, err := s.registry.LoadSchemas(ctx, userID, input.ToolNames)
		if err != nil {
			return ExploreOutput{}, err
		}
		return ExploreOutput{
			ToolSchemas: schemas,
			Message:     "Loaded requested tool schemas. Next: call execute_platform_tool with tool_name and tool_parameters.",
			NextStep: map[string]any{
				"action":     "execute_platform_tool",
				"tool_names": input.ToolNames,
			},
		}, nil
	}

	platform := strings.TrimSpace(input.Platform)
	if platform == "" {
		platforms, err := s.registry.ListPlatforms(ctx, userID)
		if err != nil {
			return ExploreOutput{}, err
		}
		return ExploreOutput{
			Platforms: platforms,
			Message:   "Found available platforms. Set platform to list tools or pass tool_names to load schemas.",
		}, nil
	}

	tools, err := s.registry.ListTools(ctx, userID, platform)
	if err != nil {
		return ExploreOutput{}, fmt.Errorf("list tools for %s: %w", platform, err)
	}

	byAction := map[string][]string{
		"read":    {},
		"propose": {},
		"execute": {},
		"other":   {},
	}
	for _, tool := range tools {
		switch tool.Action {
		case domain.ToolActionRead:
			byAction["read"] = append(byAction["read"], tool.Name)
		case domain.ToolActionPropose:
			byAction["propose"] = append(byAction["propose"], tool.Name)
		case domain.ToolActionExecute:
			byAction["execute"] = append(byAction["execute"], tool.Name)
		default:
			byAction["other"] = append(byAction["other"], tool.Name)
		}
	}

	return ExploreOutput{
		Platform:      platform,
		ToolsCount:    len(tools),
		Tools:         tools,
		ToolsByAction: byAction,
		Message:       "Found tools. Batch-load schemas in one call using tool_names before execution.",
	}, nil
}
