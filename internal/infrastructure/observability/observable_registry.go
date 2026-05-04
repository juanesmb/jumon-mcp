package observability

import (
	"context"
	"time"

	domaincatalog "jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/provider/registry"
)

// ObservableRegistry delegates catalog operations while wrapping Execute with MCP tracing and metrics.
type ObservableRegistry struct {
	inner *registry.Registry
	rec   *Recorder
}

// NewObservableRegistry constructs a decorator around the canonical tool registry.
func NewObservableRegistry(inner *registry.Registry, rec *Recorder) *ObservableRegistry {
	if inner == nil || rec == nil {
		panic("observability: ObservableRegistry requires non-nil registry and recorder")
	}
	return &ObservableRegistry{inner: inner, rec: rec}
}

func (o *ObservableRegistry) ListPlatforms(ctx context.Context, userID string) ([]domaincatalog.Platform, error) {
	return o.inner.ListPlatforms(ctx, userID)
}

func (o *ObservableRegistry) ListTools(ctx context.Context, userID, platform string) ([]domaincatalog.ToolSummary, error) {
	return o.inner.ListTools(ctx, userID, platform)
}

func (o *ObservableRegistry) LoadSchemas(ctx context.Context, userID string, toolNames []string) ([]domaincatalog.ToolSchema, error) {
	return o.inner.LoadSchemas(ctx, userID, toolNames)
}

func (o *ObservableRegistry) Execute(ctx context.Context, userID, toolName string, params map[string]any) (domaincatalog.ToolExecutionResult, error) {
	platform, _ := o.inner.PlatformForTool(toolName)

	ctxSpan, span := o.rec.StartExecuteSpan(ctx, userID, toolName, platform)

	start := time.Now()
	out, err := o.inner.Execute(ctxSpan, userID, toolName, params)

	elapsed := float64(time.Since(start).Milliseconds())
	outcome := ClassifyExecuteError(err)

	actualPlatform := platform
	if err == nil && out.ToolName != "" && actualPlatform == "" {
		actualPlatform, _ = o.inner.PlatformForTool(out.ToolName)
	}

	o.rec.RecordToolExecute(ctxSpan, toolName, actualPlatform, outcome, elapsed)

	FinishSpan(span, err)
	return out, err
}
