package googleads

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func annotateGoogleSpan(ctx context.Context, customerID, gaqlResource, mcpTool string) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}
	attrs := []attribute.KeyValue{
		attribute.String("google.mcp_tool", mcpTool),
	}
	if customerID != "" {
		attrs = append(attrs, attribute.String("google.customer_id", customerID))
	}
	if gaqlResource != "" {
		attrs = append(attrs, attribute.String("google.gaql_resource", gaqlResource))
	}
	span.SetAttributes(attrs...)
}

func logGoogleSearchFailure(ctx context.Context, mcpTool, customerID, resource, query string, err error) {
	if err == nil {
		return
	}
	attrs := []slog.Attr{
		slog.String("google.mcp_tool", mcpTool),
		slog.String("google.gaql_resource", resource),
		slog.String("error", err.Error()),
	}
	if customerID != "" {
		attrs = append(attrs, slog.String("google.customer_id", customerID))
	}
	if query != "" {
		attrs = append(attrs, slog.String("google.query", query))
	}
	slog.LogAttrs(ctx, slog.LevelInfo, "google ads search failed", attrs...)
}
