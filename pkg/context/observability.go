package context

import (
	"context"
	"time"
)

// Metrics tracks context assembly metrics.
type Metrics struct {
	AssemblyDurationMs float64
	ProviderDurations  map[string]float64
	FragmentCount      map[string]int
	TokensUsed         map[string]int
	TruncationEvents   int
	CacheHitRate       float64
	TotalTokens        int
}

// MetricsCollector collects context assembly metrics.
type MetricsCollector struct {
	metrics *Metrics
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: &Metrics{
			ProviderDurations: make(map[string]float64),
			FragmentCount:     make(map[string]int),
			TokensUsed:        make(map[string]int),
		},
	}
}

// RecordProviderDuration records provider fetch duration.
func (m *MetricsCollector) RecordProviderDuration(provider string, duration time.Duration) {
	m.metrics.ProviderDurations[provider] = float64(duration.Milliseconds())
}

// RecordFragment records a context fragment.
func (m *MetricsCollector) RecordFragment(fragmentType string, tokens int) {
	m.metrics.FragmentCount[fragmentType]++
	m.metrics.TokensUsed[fragmentType] += tokens
	m.metrics.TotalTokens += tokens
}

// RecordTruncation records a truncation event.
func (m *MetricsCollector) RecordTruncation() {
	m.metrics.TruncationEvents++
}

// RecordAssemblyDuration records total assembly duration.
func (m *MetricsCollector) RecordAssemblyDuration(duration time.Duration) {
	m.metrics.AssemblyDurationMs = float64(duration.Milliseconds())
}

// GetMetrics returns the collected metrics.
func (m *MetricsCollector) GetMetrics() *Metrics {
	return m.metrics
}

// TracingSpan represents a tracing span for context assembly.
type TracingSpan struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Metadata  map[string]any
}

// Tracer manages tracing spans.
type Tracer struct {
	spans []TracingSpan
}

// NewTracer creates a new tracer.
func NewTracer() *Tracer {
	return &Tracer{
		spans: []TracingSpan{},
	}
}

// StartSpan starts a new tracing span.
func (t *Tracer) StartSpan(ctx context.Context, name string) *TracingSpan {
	span := &TracingSpan{
		Name:      name,
		StartTime: time.Now(),
		Metadata:  make(map[string]any),
	}
	return span
}

// EndSpan ends a tracing span.
func (t *Tracer) EndSpan(span *TracingSpan) {
	span.EndTime = time.Now()
	t.spans = append(t.spans, *span)
}

// GetSpans returns all tracing spans.
func (t *Tracer) GetSpans() []TracingSpan {
	return t.spans
}

// ObservableEngine wraps ContextEngine with observability.
type ObservableEngine struct {
	Engine  *ContextEngine
	Metrics *MetricsCollector
	Tracer  *Tracer
}

// NewObservableEngine creates a new observable engine.
func NewObservableEngine(engine *ContextEngine) *ObservableEngine {
	return &ObservableEngine{
		Engine:  engine,
		Metrics: NewMetricsCollector(),
		Tracer:  NewTracer(),
	}
}

// FetchAll wraps engine FetchAll with metrics.
func (o *ObservableEngine) FetchAll(ctx context.Context, input ContextInput) ([]ContextFragment, error) {
	span := o.Tracer.StartSpan(ctx, "context.assemble")
	defer o.Tracer.EndSpan(span)

	startTime := time.Now()

	// Track individual provider fetches
	fragments := []ContextFragment{}
	for _, provider := range o.Engine.Providers {
		providerSpan := o.Tracer.StartSpan(ctx, "context.provider.fetch")
		providerSpan.Metadata["provider"] = provider.Name()

		providerStart := time.Now()
		fragment, err := provider.Fetch(ctx, input)
		providerDuration := time.Since(providerStart)

		o.Metrics.RecordProviderDuration(provider.Name(), providerDuration)
		o.Tracer.EndSpan(providerSpan)

		if err == nil && fragment != nil {
			fragments = append(fragments, *fragment)
			o.Metrics.RecordFragment(string(fragment.Type), fragment.TokenCount)
		}
	}

	duration := time.Since(startTime)
	o.Metrics.RecordAssemblyDuration(duration)

	return fragments, nil
}

// Assemble wraps engine Assemble with metrics.
func (o *ObservableEngine) Assemble(fragments []ContextFragment, config AssemblyConfig) ([]any, error) {
	span := o.Tracer.StartSpan(context.Background(), "context.token_budget")
	defer o.Tracer.EndSpan(span)

	// This is a placeholder - actual implementation would call engine.Assemble
	// and track truncation events

	return nil, nil
}
