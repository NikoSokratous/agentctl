package observe

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("agentruntime/observe")

// OTELExporter exports traces to OpenTelemetry.
type OTELExporter struct {
	tp *sdktrace.TracerProvider
}

// NewOTELExporter creates an exporter that sends traces to an OTLP endpoint.
func NewOTELExporter(ctx context.Context, endpoint string) (*OTELExporter, error) {
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("agentruntime"),
		)),
	)

	return &OTELExporter{tp: tp}, nil
}

// Shutdown flushes and stops the tracer provider.
func (e *OTELExporter) Shutdown(ctx context.Context) error {
	return e.tp.Shutdown(ctx)
}

// StartSpan creates a span for an execution step.
func (e *OTELExporter) StartSpan(ctx context.Context, runID, stepID, name string, attrs ...attribute.KeyValue) (context.Context, func()) {
	attrs = append(attrs,
		attribute.String("run_id", runID),
		attribute.String("step_id", stepID),
	)
	ctx, span := tracer.Start(ctx, name, trace.WithAttributes(attrs...))
	return ctx, func() { span.End() }
}

// RecordStep records a step as a span with reasoning summary.
func (e *OTELExporter) RecordStep(ctx context.Context, runID string, evt Event) (context.Context, func()) {
	name := string(evt.Type)
	if evt.Type == EventToolCall {
		if t, ok := evt.Data["tool"].(string); ok {
			name = "tool:" + t
		}
	}
	attrs := []attribute.KeyValue{
		attribute.String("run_id", evt.RunID),
		attribute.String("step_id", evt.StepID),
		attribute.String("agent", evt.Agent),
	}
	if r, ok := evt.Data["reasoning"].(string); ok {
		attrs = append(attrs, attribute.String("reasoning", r))
	}
	if tool, ok := evt.Data["tool"].(string); ok {
		attrs = append(attrs, attribute.String("tool", tool))
	}

	ctx, span := tracer.Start(ctx, name, trace.WithTimestamp(evt.Timestamp), trace.WithAttributes(attrs...))
	return ctx, func() { span.End() }
}

// Export sends events to the OTLP endpoint as spans.
func (e *OTELExporter) Export(ctx context.Context, events []Event) error {
	for _, evt := range events {
		_, end := e.RecordStep(ctx, evt.RunID, evt)
		end()
	}
	return nil
}
