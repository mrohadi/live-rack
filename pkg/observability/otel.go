// Package observability bootstraps OpenTelemetry tracing + metrics for live-rack services.
// Each service calls Setup at startup and defers the returned shutdown func.
package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds OTel bootstrap options.
type Config struct {
	ServiceName    string
	ServiceVersion string
	// OTLPEndpoint is the gRPC endpoint for Tempo (e.g. "localhost:4317").
	// If empty, tracing is disabled.
	OTLPEndpoint string
}

// Provider bundles the tracer and meter providers.
type Provider struct {
	tp *sdktrace.TracerProvider
	mp *sdkmetric.MeterProvider
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer { return otel.Tracer(name) }

// Meter returns a named meter from the global provider.
func Meter(name string) metric.Meter { return otel.Meter(name) }

// Setup initialises OTel with OTLP trace export to Tempo and Prometheus metric export.
// Returns a shutdown func that must be deferred.
func Setup(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource merge: %w", err)
	}

	var shutdowns []func(context.Context) error

	// Tracing — OTLP gRPC to Tempo.
	if cfg.OTLPEndpoint != "" {
		conn, err := grpc.NewClient(cfg.OTLPEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("dial otlp: %w", err)
		}
		traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("trace exporter: %w", err)
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExp),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		otel.SetTracerProvider(tp)
		shutdowns = append(shutdowns, tp.Shutdown)
	}

	// Metrics — Prometheus pull endpoint.
	promExp, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("prometheus exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExp),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	shutdowns = append(shutdowns, mp.Shutdown)

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		for _, fn := range shutdowns {
			if err := fn(ctx); err != nil {
				return err
			}
		}
		return nil
	}, nil
}
