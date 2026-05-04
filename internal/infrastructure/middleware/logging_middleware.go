package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	stdtrace "go.opentelemetry.io/otel/trace"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// LoggingHandler emits structured inbound HTTP logs with trace correlation compatible with GCP Cloud Logging JSON fields.
func LoggingHandler(handler http.Handler, gcpProjectID string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		handler.ServeHTTP(wrapped, r)

		durationMs := float64(time.Since(start).Milliseconds())

		event := "http_request"
		attrs := []slog.Attr{
			slog.String("event", event),
			slog.String("http.method", r.Method),
			slog.String("http.path", r.URL.Path),
			slog.Int("status", wrapped.statusCode),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Float64("duration_ms", durationMs),
		}
		attrs = appendGCPTrace(attrs, r.Context(), gcpProjectID)

		slog.LogAttrs(r.Context(), slog.LevelInfo, event, attrs...)
	})
}

func appendGCPTrace(attrs []slog.Attr, ctx context.Context, projectID string) []slog.Attr {
	reqCtx := stdtrace.SpanContextFromContext(ctx)
	if !reqCtx.TraceID().IsValid() {
		return attrs
	}
	attrs = append(attrs, slog.String("trace_id", reqCtx.TraceID().String()))
	if projectID != "" {
		traceField := "projects/" + projectID + "/traces/" + reqCtx.TraceID().String()
		attrs = append(attrs, slog.String("logging.googleapis.com/trace", traceField))
	}
	if reqCtx.SpanID().IsValid() {
		attrs = append(attrs, slog.String("logging.googleapis.com/spanId", reqCtx.SpanID().String()))
	}
	return attrs
}
