package observability

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware provides tracing for HTTP requests.
func HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from headers
			ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span
			ctx, span := StartSpan(ctx, serviceName, r.Method+" "+r.URL.Path,
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.user_agent", r.Header.Get("User-Agent")),
			)
			defer span.End()

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Execute request
			start := time.Now()
			next.ServeHTTP(rw, r.WithContext(ctx))
			duration := time.Since(start)

			// Record response attributes
			span.SetAttributes(
				attribute.Int("http.status_code", rw.statusCode),
				attribute.Int64("http.response_size", rw.bytesWritten),
				attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
			)

			// Record error if status code indicates failure
			if rw.statusCode >= 400 {
				span.SetAttributes(attribute.Bool("error", true))
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// TraceWorkflow traces workflow execution.
func TraceWorkflow(ctx context.Context, workflowName string) (context.Context, trace.Span) {
	return StartSpan(ctx, "workflow", "workflow.execute",
		attribute.String("workflow.name", workflowName),
	)
}

// TraceWorkflowStep traces a workflow step.
func TraceWorkflowStep(ctx context.Context, stepName, agentName string) (context.Context, trace.Span) {
	return StartSpan(ctx, "workflow", "workflow.step",
		attribute.String("step.name", stepName),
		attribute.String("step.agent", agentName),
	)
}

// TraceToolExecution traces tool execution.
func TraceToolExecution(ctx context.Context, toolName, action string) (context.Context, trace.Span) {
	return StartSpan(ctx, "tool", "tool.execute",
		attribute.String("tool.name", toolName),
		attribute.String("tool.action", action),
	)
}

// TracePolicyEvaluation traces policy evaluation.
func TracePolicyEvaluation(ctx context.Context, policyName string) (context.Context, trace.Span) {
	return StartSpan(ctx, "policy", "policy.evaluate",
		attribute.String("policy.name", policyName),
	)
}

// TraceRiskAssessment traces risk assessment.
func TraceRiskAssessment(ctx context.Context, action string) (context.Context, trace.Span) {
	return StartSpan(ctx, "risk", "risk.assess",
		attribute.String("risk.action", action),
	)
}

// TraceDatabase traces database operations.
func TraceDatabase(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	return StartSpan(ctx, "database", "db."+operation,
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	)
}

// PropagationMiddleware propagates trace context.
func PropagationMiddleware() func(http.Handler) http.Handler {
	propagator := propagation.TraceContext{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// InjectTraceContext injects trace context into HTTP headers.
func InjectTraceContext(ctx context.Context, req *http.Request) {
	propagator := propagation.TraceContext{}
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}
