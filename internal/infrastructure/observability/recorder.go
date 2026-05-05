package observability

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"jumon-mcp/internal/config"
	"jumon-mcp/internal/domain/catalog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	namespace = "mcp"

	instrExecuteCalls          = namespace + "_tool_execute_calls_total"
	instrExploreCalls          = namespace + "_explore_calls_total"
	instrToolExecuteDurationMs = namespace + "_tool_execute_duration_ms"
	instrExploreDurationMs     = namespace + "_explore_duration_ms"
	instrUpstreamRequests      = namespace + "_http_upstream_requests_total"
	instrUpstreamDurationMs    = namespace + "_http_upstream_duration_ms"
	instrPlatformNotConnected  = namespace + "_platform_not_connected_total"
)

// Recorder holds OpenTelemetry instruments and emits structured telemetry for MCP tooling and outbound HTTP.
type Recorder struct {
	cfg    config.ObservabilityConfig
	tracer trace.Tracer

	execCalls          metric.Int64Counter
	exploreCalls       metric.Int64Counter
	toolExecDurationMs metric.Float64Histogram
	exploreDurationMs  metric.Float64Histogram
	upstreamRequests   metric.Int64Counter
	upstreamDurMs      metric.Float64Histogram
	platformNotConn    metric.Int64Counter
}

// NewRecorder wires metric instruments backed by OpenTelemetry Meter. When noop is true, Recording methods no-op safely.
func NewRecorder(cfg config.ObservabilityConfig, m metric.Meter, t trace.Tracer, noopInstruments bool) (*Recorder, error) {
	rec := &Recorder{
		cfg:    cfg,
		tracer: t,
	}
	if noopInstruments {
		return rec, nil
	}

	var err error
	rec.execCalls, err = m.Int64Counter(instrExecuteCalls)
	if err != nil {
		return nil, err
	}
	rec.exploreCalls, err = m.Int64Counter(instrExploreCalls)
	if err != nil {
		return nil, err
	}
	rec.toolExecDurationMs, err = m.Float64Histogram(
		instrToolExecuteDurationMs,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of MCP execute_platform_tool calls"),
	)
	if err != nil {
		return nil, err
	}
	rec.exploreDurationMs, err = m.Float64Histogram(
		instrExploreDurationMs,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of MCP explore_platform calls"),
	)
	if err != nil {
		return nil, err
	}
	rec.upstreamRequests, err = m.Int64Counter(instrUpstreamRequests)
	if err != nil {
		return nil, err
	}
	rec.upstreamDurMs, err = m.Float64Histogram(
		instrUpstreamDurationMs,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of outbound gateway HTTP calls"),
	)
	if err != nil {
		return nil, err
	}
	rec.platformNotConn, err = m.Int64Counter(instrPlatformNotConnected)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (r *Recorder) Enabled() bool { return r.cfg.Enabled }

func (r *Recorder) Tracer() trace.Tracer { return r.tracer }

func (r *Recorder) GCPProjectID() string { return r.cfg.GCPProjectID }

func (r *Recorder) UserHash(userID string) string {
	return UserHashHex(userID, r.cfg.UserIDHashSalt)
}

// ExploreMode derives a coarse explore phase label for dashboards.
func ExploreMode(platform string, toolNames []string) string {
	switch {
	case len(toolNames) > 0:
		return "load_schemas"
	case strings.TrimSpace(platform) == "":
		return "list_platforms"
	default:
		return "list_tools"
	}
}

func (r *Recorder) RecordExplore(ctx context.Context, exploreMode string, outcome string, durationMs float64) {
	if r.exploreCalls != nil {
		r.exploreCalls.Add(ctx, 1,
			metric.WithAttributes(attribute.String("explore_mode", exploreMode), attribute.String("outcome", outcome)))
	}
	if r.exploreDurationMs != nil && durationMs >= 0 {
		r.exploreDurationMs.Record(ctx, durationMs,
			metric.WithAttributes(attribute.String("explore_mode", exploreMode), attribute.String("outcome", outcome)))
	}
}

func (r *Recorder) RecordToolExecute(ctx context.Context, toolName, platform, outcome string, durationMs float64) {
	if r.execCalls != nil {
		attrs := []attribute.KeyValue{attribute.String("tool_name", toolName), attribute.String("outcome", outcome)}
		if platform != "" {
			attrs = append(attrs, attribute.String("platform", platform))
		}
		r.execCalls.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
	if r.toolExecDurationMs != nil && durationMs >= 0 {
		attrs := []attribute.KeyValue{attribute.String("tool_name", toolName), attribute.String("outcome", outcome)}
		if platform != "" {
			attrs = append(attrs, attribute.String("platform", platform))
		}
		r.toolExecDurationMs.Record(ctx, durationMs, metric.WithAttributes(attrs...))
	}
}

func (r *Recorder) RecordUpstreamHTTP(ctx context.Context, method string, routeTemplate string, provider string, status int, retry bool, attempt int, durationMs float64, errMsg string) {
	sclass := StatusClass(status)
	method = NormalizeHTTPMethod(method)
	event := "http_upstream"

	proj := r.cfg.GCPProjectID

	var traceLog string
	if sp := trace.SpanContextFromContext(ctx); sp.TraceID().IsValid() && proj != "" {
		traceLog = "projects/" + proj + "/traces/" + sp.TraceID().String()
	}

	attrs := []slog.Attr{
		slog.String("event", event),
		slog.String("http.method", method),
		slog.String("http.route", routeTemplate),
		slog.String("gateway.provider", provider),
		slog.Int("http.status_code", status),
		slog.String("http.status_class", sclass),
		slog.Bool("retry", retry),
		slog.Int("attempt", attempt),
		slog.Float64("duration_ms", durationMs),
	}
	if traceLog != "" {
		attrs = append(attrs, slog.String("logging.googleapis.com/trace", traceLog))
	}
	if sp := trace.SpanContextFromContext(ctx); sp.SpanID().IsValid() {
		attrs = append(attrs, slog.String("logging.googleapis.com/spanId", sp.SpanID().String()))
	}
	if errMsg != "" {
		attrs = append(attrs, slog.String("error_message", errMsg))
	}
	slog.LogAttrs(ctx, slog.LevelInfo, event, attrs...)

	if r.upstreamRequests != nil {
		r.upstreamRequests.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("gateway.provider", provider),
				attribute.String("http.method", method),
				attribute.String("http.route", routeTemplate),
				attribute.String("http.status_class", sclass),
			))
	}
	if r.upstreamDurMs != nil && durationMs >= 0 {
		r.upstreamDurMs.Record(ctx, durationMs,
			metric.WithAttributes(
				attribute.String("gateway.provider", provider),
				attribute.String("http.method", method),
				attribute.String("http.route", routeTemplate),
				attribute.String("http.status_class", sclass),
			))
	}
}

func (r *Recorder) RecordPlatformNotConnected(ctx context.Context, platform string) {
	if r.platformNotConn != nil {
		r.platformNotConn.Add(ctx, 1, metric.WithAttributes(attribute.String("platform", platform)))
	}
}

// StartExploreSpan starts a span for explore_platform; caller must end the span and set status.
func (r *Recorder) StartExploreSpan(ctx context.Context, userID, exploreMode string) (context.Context, trace.Span) {
	return r.tracer.Start(ctx, "mcp.tool.explore",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("mcp.explore_mode", exploreMode),
			attribute.String("user.hash", r.UserHash(userID)),
		))
}

// FinishSpan records error state on the span and ends it.
func FinishSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// StartExecuteSpan starts a span for execute handling at the registry boundary.
func (r *Recorder) StartExecuteSpan(ctx context.Context, userID, toolName, platform string) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("tool.name", toolName),
			attribute.String("user.hash", r.UserHash(userID)),
		),
	}
	if platform != "" {
		opts = append(opts, trace.WithAttributes(attribute.String("platform", platform)))
	}
	return r.tracer.Start(ctx, "mcp.tool.execute", opts...)
}

func ClassifyExecuteError(err error) string {
	if err == nil {
		return "success"
	}
	var notConn *catalog.PlatformNotConnectedError
	if errors.As(err, &notConn) {
		return "platform_not_connected"
	}
	return "error"
}
