package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Propagator manages trace context propagation.
type Propagator struct {
	propagator propagation.TextMapPropagator
}

// NewPropagator creates a new trace propagator.
func NewPropagator() *Propagator {
	return &Propagator{
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}
}

// Extract extracts trace context from headers.
func (p *Propagator) Extract(ctx context.Context, headers http.Header) context.Context {
	return p.propagator.Extract(ctx, propagation.HeaderCarrier(headers))
}

// Inject injects trace context into headers.
func (p *Propagator) Inject(ctx context.Context, headers http.Header) {
	p.propagator.Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractFromMap extracts trace context from a map.
func (p *Propagator) ExtractFromMap(ctx context.Context, carrier map[string]string) context.Context {
	return p.propagator.Extract(ctx, propagation.MapCarrier(carrier))
}

// InjectToMap injects trace context into a map.
func (p *Propagator) InjectToMap(ctx context.Context, carrier map[string]string) {
	p.propagator.Inject(ctx, propagation.MapCarrier(carrier))
}

// GetSpanContext retrieves the current span context.
func GetSpanContext(ctx context.Context) trace.SpanContext {
	return trace.SpanFromContext(ctx).SpanContext()
}

// GetTraceID retrieves the current trace ID.
func GetTraceID(ctx context.Context) string {
	spanCtx := GetSpanContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID retrieves the current span ID.
func GetSpanID(ctx context.Context) string {
	spanCtx := GetSpanContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// IsTracing checks if tracing is active in the context.
func IsTracing(ctx context.Context) bool {
	return trace.SpanFromContext(ctx).SpanContext().IsValid()
}

// WithTraceContext creates a new context with trace context from headers.
func WithTraceContext(ctx context.Context, headers http.Header) context.Context {
	propagator := NewPropagator()
	return propagator.Extract(ctx, headers)
}
