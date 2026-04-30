package execution

import (
	"context"
	"fmt"
	"strings"

	"jumon-mcp/internal/domain/catalog"
)

type Registry interface {
	Execute(ctx context.Context, userID, toolName string, params map[string]any) (catalog.ToolExecutionResult, error)
}

type Service struct {
	registry Registry
}

func NewService(registry Registry) *Service {
	return &Service{registry: registry}
}

type ExecuteInput struct {
	ToolName       string
	ToolParameters map[string]any
}

type ExecuteOutput struct {
	Result catalog.ToolExecutionResult `json:"result"`
}

func (s *Service) Execute(ctx context.Context, userID string, input ExecuteInput) (ExecuteOutput, error) {
	name := strings.TrimSpace(input.ToolName)
	if name == "" {
		return ExecuteOutput{}, fmt.Errorf("tool_name is required")
	}
	parameters := input.ToolParameters
	if parameters == nil {
		parameters = map[string]any{}
	}
	result, err := s.registry.Execute(ctx, userID, name, parameters)
	if err != nil {
		return ExecuteOutput{}, err
	}
	return ExecuteOutput{Result: result}, nil
}
