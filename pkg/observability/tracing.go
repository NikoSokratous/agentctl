package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig defines tracing configuration.
type TracingConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Exporter    string  `yaml:"exporter"` // otlp, zipkin, stdout
	Endpoint    string  `yaml:"endpoint"` // e.g., localhost:4317
	ServiceName string  `yaml:"service_name"`
	SampleRate  float64 `yaml:"sample_rate"` // 0.0 to 1.0
}

// TracingProvider manages OpenTelemetry tracing.
type TracingProvider struct {
	provider *sdktrace.TracerProvider
	config   *TracingConfig
}

// NewTracingProvider creates a new tracing provider.
func NewTracingProvider(config *TracingConfig) (*TracingProvider, error) {
	if !config.Enabled {
		return &TracingProvider{config: config}, nil
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String("0.6.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	// Create exporter based on configuration
	var exporter sdktrace.SpanExporter
	switch config.Exporter {
	case "otlp":
		exporter, err = createOTLPExporter(config.Endpoint)
	case "zipkin":
		exporter, err = createZipkinExporter(config.Endpoint)
	case "stdout":
		exporter, err = createStdoutExporter()
	default:
		return nil, fmt.Errorf("unsupported exporter: %s", config.Exporter)
	}

	if err != nil {
		return nil, fmt.Errorf("create exporter: %w", err)
	}

	// Create sampler
	sampler := sdktrace.AlwaysSample()
	if config.SampleRate < 1.0 {
		sampler = sdktrace.TraceIDRatioBased(config.SampleRate)
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set as global provider
	otel.SetTracerProvider(provider)

	return &TracingProvider{
		provider: provider,
		config:   config,
	}, nil
}

// Shutdown gracefully shuts down the tracing provider.
func (tp *TracingProvider) Shutdown(ctx context.Context) error {
	if tp.provider != nil {
		return tp.provider.Shutdown(ctx)
	}
	return nil
}

// GetTracer returns a named tracer.
func (tp *TracingProvider) GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// createOTLPExporter creates an OTLP exporter.
func createOTLPExporter(endpoint string) (sdktrace.SpanExporter, error) {
	ctx := context.Background()

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	return exporter, nil
}

// createZipkinExporter creates a Zipkin exporter.
func createZipkinExporter(endpoint string) (sdktrace.SpanExporter, error) {
	exporter, err := zipkin.New(endpoint)
	if err != nil {
		return nil, fmt.Errorf("create Zipkin exporter: %w", err)
	}

	return exporter, nil
}

// createStdoutExporter creates a stdout exporter (for development).
func createStdoutExporter() (sdktrace.SpanExporter, error) {
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, fmt.Errorf("create stdout exporter: %w", err)
	}

	return exporter, nil
}

// StartSpan starts a new span with common attributes.
func StartSpan(ctx context.Context, tracerName, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	return tracer.Start(ctx, spanName, trace.WithAttributes(attrs...))
}

// RecordError records an error in the current span.
func RecordError(ctx context.Context, err error) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.RecordError(err)
	}
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetAttributes(attrs...)
	}
}

// AddEvent adds an event to the current span.
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}
