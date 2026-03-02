package orchestrate

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RunsCreated tracks number of runs created
	RunsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agentruntime_runs_created_total",
		Help: "Total number of runs created",
	})

	// RunsActive tracks number of currently active runs
	RunsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "agentruntime_runs_active",
		Help: "Number of currently active runs",
	})

	// RunDuration tracks run completion time
	RunDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "agentruntime_run_duration_seconds",
		Help:    "Duration of run execution in seconds",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	})

	// APIRequestDuration tracks API request duration
	APIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "agentruntime_api_request_duration_seconds",
		Help:    "API request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint", "status"})

	// ToolExecutions tracks tool execution counts
	ToolExecutions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentruntime_tool_executions_total",
		Help: "Total number of tool executions",
	}, []string{"tool", "status"})

	// PolicyDenials tracks policy denials
	PolicyDenials = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentruntime_policy_denials_total",
		Help: "Total number of policy denials",
	}, []string{"rule"})
)
