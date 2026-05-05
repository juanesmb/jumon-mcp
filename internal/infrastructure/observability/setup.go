package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"jumon-mcp/internal/config"

	gcpdetect "go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	mopexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Setup configures global OpenTelemetry providers and constructs a Recorder. When cfg.Enabled is false, only no-op exporters are wired.
//
// Returned shutdown MUST be invoked so span batches and periodic metric readers flush.
func Setup(ctx context.Context, cfg config.ObservabilityConfig) (_ *Recorder, shutdown func(context.Context) error, err error) {
	noopShutdown := func(context.Context) error { return nil }

	if !cfg.Enabled {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
		otel.SetMeterProvider(metricnoop.NewMeterProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		rec, err := NewRecorder(cfg, otel.GetMeterProvider().Meter("mcp.metrics"), otel.GetTracerProvider().Tracer(cfg.ServiceName), true)
		if err != nil {
			return nil, noopShutdown, err
		}
		return rec, noopShutdown, nil
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithDetectors(gcpdetect.NewDetector()),
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil && !errors.Is(err, resource.ErrPartialResource) {
		return nil, noopShutdown, fmt.Errorf("build telemetry resource: %w", err)
	}

	traceOpts := []texporter.Option{}
	metricOpts := []mopexporter.Option{}
	if cfg.GCPProjectID != "" {
		traceOpts = append(traceOpts, texporter.WithProjectID(cfg.GCPProjectID))
		metricOpts = append(metricOpts, mopexporter.WithProjectID(cfg.GCPProjectID))
	}

	texp, err := texporter.New(traceOpts...)
	if err != nil {
		return nil, noopShutdown, fmt.Errorf("cloud trace exporter: %w", err)
	}

	ratio := cfg.TraceSampleRatio
	switch {
	case ratio < 0:
		ratio = 0
	case ratio > 1:
		ratio = 1
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(texp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
	otel.SetTracerProvider(tp)

	mexp, err := mopexporter.New(metricOpts...)
	if err != nil {
		_ = tp.Shutdown(context.Background())
		return nil, noopShutdown, fmt.Errorf("cloud monitoring metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(mexp)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	otel.SetMeterProvider(mp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	rec, err := NewRecorder(cfg, otel.GetMeterProvider().Meter("mcp.metrics"), tp.Tracer(cfg.ServiceName), false)
	if err != nil {
		shErr := errors.Join(mp.Shutdown(ctx), tp.Shutdown(ctx))
		if shErr != nil {
			slog.Error("telemetry shutdown failed", "error", shErr.Error())
		}
		return nil, noopShutdown, err
	}

	shutdownFn := func(c context.Context) error {
		errFirst := tp.Shutdown(c)
		if errShutdown := mp.Shutdown(c); errShutdown != nil {
			return errors.Join(errFirst, errShutdown)
		}
		return errFirst
	}

	return rec, shutdownFn, nil
}
