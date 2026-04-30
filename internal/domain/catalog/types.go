package catalog

import "fmt"

type ToolAction string

const (
	ToolActionRead    ToolAction = "read"
	ToolActionPropose ToolAction = "propose"
	ToolActionExecute ToolAction = "execute"
	ToolActionOther   ToolAction = "other"
)

type Platform struct {
	Name       string `json:"name"`
	Connected  bool   `json:"connected"`
	ConnectURL string `json:"connect_url,omitempty"`
}

type ToolSummary struct {
	Name     string     `json:"name"`
	Platform string     `json:"platform"`
	Action   ToolAction `json:"action"`
	Summary  string     `json:"summary"`
}

type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type ToolExecutionResult struct {
	ToolName string `json:"tool_name"`
	Data     any    `json:"data"`
}

type PlatformNotConnectedError struct {
	Platform   string `json:"platform"`
	ConnectURL string `json:"connect_url"`
}

func (e *PlatformNotConnectedError) Error() string {
	return fmt.Sprintf("%s is not connected", e.Platform)
}
