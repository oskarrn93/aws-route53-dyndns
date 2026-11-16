package telemetry

import (
	"aws-route53-dyndns/internal/config"
	"context"
	"fmt"
	"log/slog"

	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Metric struct {
	Name        string
	Unit        string
	Description string
}

var SuccessfulRunsMetric = Metric{
	Name:        "successful_runs",
	Description: "Number of successful runs",
}

var FailedRunsMetric = Metric{
	Name:        "failed_runs",
	Description: "Number of faileds runs",
}

var IPAddressChangedMetric = Metric{
	Name:        "ip_address_changed",
	Description: "The ip address has changed",
}

var IPAddressNotChangedMetric = Metric{
	Name:        "ip_address_not_changed",
	Description: "The ip address has not changed",
}

// // TelemetryProvider is an interface for the telemetry provider.
// type TelemetryProvider interface {
// 	GetServiceName() string
// 	MeterInt64Histogram(metric Metric) (otelmetric.Int64Histogram, error)
// 	MeterInt64UpDownCounter(metric Metric) (otelmetric.Int64UpDownCounter, error)
// 	TraceStart(ctx context.Context, name string) (context.Context, oteltrace.Span)
// 	Shutdown(ctx context.Context)
// }

// Telemetry is a wrapper around the OpenTelemetry logger, meter, and tracer.
type Telemetry struct {
	LoggerProvider *log.LoggerProvider
	MeterProvider  *metric.MeterProvider
	TracerProvider *trace.TracerProvider
	Logger         *slog.Logger
	Meter          otelmetric.Meter
	Tracer         oteltrace.Tracer
	Config         *config.Config
}

func (t *Telemetry) Increment(ctx context.Context, metric Metric) {
	counter, err := t.Meter.Int64Counter(
		metric.Name,
		otelmetric.WithDescription(metric.Description),
		otelmetric.WithUnit(metric.Unit),
	)
	if err != nil {
		t.Logger.Warn("failed to initializer metric counter", "error", err)
	}

	counter.Add(ctx, 1)

}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Telemetry, error) {

	rp := newResource(cfg.ServiceName, cfg.ServiceVersion)

	loggerProvider, err := newLoggerProvider(ctx, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	meterProvider, err := newMeterProvider(ctx, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter: %w", err)
	}
	meter := meterProvider.Meter(cfg.ServiceName)

	tracerProvider, err := newTracerProvider(ctx, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)

	return &Telemetry{
		LoggerProvider: loggerProvider,
		MeterProvider:  meterProvider,
		TracerProvider: tracerProvider,
		Logger:         logger,
		Meter:          meter,
		Tracer:         tracer,
		Config:         cfg,
	}, nil
}

// MeterInt64Histogram creates a new int64 histogram metric.
func (t *Telemetry) MeterInt64Histogram(metric Metric) (otelmetric.Int64Histogram, error) {
	histogram, err := t.Meter.Int64Histogram(
		metric.Name,
		otelmetric.WithDescription(metric.Description),
		otelmetric.WithUnit(metric.Unit),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create histogram: %w", err)
	}

	return histogram, nil
}

// MeterInt64UpDownCounter creates a new int64 up down counter metric.
func (t *Telemetry) MeterInt64UpDownCounter(metric Metric) (otelmetric.Int64UpDownCounter, error) {
	counter, err := t.Meter.Int64UpDownCounter(
		metric.Name,
		otelmetric.WithDescription(metric.Description),
		otelmetric.WithUnit(metric.Unit),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create counter: %w", err)
	}

	return counter, nil
}

// TraceStart starts a new span with the given name. The span must be ended by calling End.
func (t *Telemetry) TraceStart(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	//nolint: spancheck
	return t.Tracer.Start(ctx, name)
}

// Shutdown shuts down the logger, meter, and tracer.
func (t *Telemetry) Shutdown(ctx context.Context) {
	t.LoggerProvider.Shutdown(ctx)
	t.MeterProvider.Shutdown(ctx)
	t.TracerProvider.Shutdown(ctx)
}
