package autoscaling

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutoScaler handles automatic scaling based on custom metrics
type AutoScaler struct {
	policies      map[string]*ScalingPolicy
	metrics       MetricsProvider
	scaler        ScalerBackend
	mu            sync.RWMutex
	checkInterval time.Duration
	stopChan      chan struct{}
}

// ScalingPolicy defines when and how to scale
type ScalingPolicy struct {
	ID                 string
	ServiceID          string
	TenantID           string
	MinReplicas        int
	MaxReplicas        int
	ScaleUpThreshold   float64
	ScaleDownThreshold float64
	Metrics            []MetricConfig
	CooldownPeriod     time.Duration
	LastScaleTime      time.Time
	Enabled            bool
}

// MetricConfig defines a metric to monitor for scaling
type MetricConfig struct {
	Name      string
	Type      MetricType
	Threshold float64
	Weight    float64 // For weighted average of multiple metrics
}

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCPU               MetricType = "cpu"
	MetricTypeMemory            MetricType = "memory"
	MetricTypeQueueDepth        MetricType = "queue_depth"
	MetricTypeRequestRate       MetricType = "request_rate"
	MetricTypeLatency           MetricType = "latency"
	MetricTypeErrorRate         MetricType = "error_rate"
	MetricTypeActiveConnections MetricType = "active_connections"
	MetricTypeCustom            MetricType = "custom"
)

// MetricsProvider provides metric values
type MetricsProvider interface {
	GetMetric(ctx context.Context, serviceID, tenantID string, metricType MetricType) (float64, error)
}

// ScalerBackend performs the actual scaling
type ScalerBackend interface {
	Scale(ctx context.Context, serviceID, tenantID string, replicas int) error
	GetCurrentReplicas(ctx context.Context, serviceID, tenantID string) (int, error)
}

// NewAutoScaler creates a new auto scaler
func NewAutoScaler(metrics MetricsProvider, scaler ScalerBackend) *AutoScaler {
	return &AutoScaler{
		policies:      make(map[string]*ScalingPolicy),
		metrics:       metrics,
		scaler:        scaler,
		checkInterval: 30 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the auto scaler
func (as *AutoScaler) Start(ctx context.Context) {
	go as.scalingLoop(ctx)
}

// Stop stops the auto scaler
func (as *AutoScaler) Stop() {
	close(as.stopChan)
}

// AddPolicy adds a scaling policy
func (as *AutoScaler) AddPolicy(policy *ScalingPolicy) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.policies[policy.ID] = policy
}

// RemovePolicy removes a scaling policy
func (as *AutoScaler) RemovePolicy(policyID string) {
	as.mu.Lock()
	defer as.mu.Unlock()
	delete(as.policies, policyID)
}

// GetPolicy returns a scaling policy
func (as *AutoScaler) GetPolicy(policyID string) (*ScalingPolicy, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	policy, exists := as.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policyID)
	}
	return policy, nil
}

// scalingLoop continuously checks and applies scaling policies
func (as *AutoScaler) scalingLoop(ctx context.Context) {
	ticker := time.NewTicker(as.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopChan:
			return
		case <-ticker.C:
			as.checkAndScale(ctx)
		}
	}
}

// checkAndScale checks all policies and scales if needed
func (as *AutoScaler) checkAndScale(ctx context.Context) {
	as.mu.RLock()
	policies := make([]*ScalingPolicy, 0, len(as.policies))
	for _, policy := range as.policies {
		if policy.Enabled {
			policies = append(policies, policy)
		}
	}
	as.mu.RUnlock()

	for _, policy := range policies {
		if err := as.evaluatePolicy(ctx, policy); err != nil {
			fmt.Printf("Error evaluating policy %s: %v\n", policy.ID, err)
		}
	}
}

// evaluatePolicy evaluates a single policy and scales if needed
func (as *AutoScaler) evaluatePolicy(ctx context.Context, policy *ScalingPolicy) error {
	// Check cooldown period
	if time.Since(policy.LastScaleTime) < policy.CooldownPeriod {
		return nil
	}

	// Get current replicas
	currentReplicas, err := as.scaler.GetCurrentReplicas(ctx, policy.ServiceID, policy.TenantID)
	if err != nil {
		return fmt.Errorf("get current replicas: %w", err)
	}

	// Calculate weighted metric score
	totalScore := 0.0
	totalWeight := 0.0

	for _, metricConfig := range policy.Metrics {
		value, err := as.metrics.GetMetric(ctx, policy.ServiceID, policy.TenantID, metricConfig.Type)
		if err != nil {
			fmt.Printf("Failed to get metric %s: %v\n", metricConfig.Name, err)
			continue
		}

		// Normalize to 0-1 scale based on threshold
		normalizedValue := value / metricConfig.Threshold
		totalScore += normalizedValue * metricConfig.Weight
		totalWeight += metricConfig.Weight
	}

	if totalWeight == 0 {
		return fmt.Errorf("no valid metrics")
	}

	avgScore := totalScore / totalWeight

	// Determine scaling action
	var desiredReplicas int
	scaleAction := ""

	if avgScore > policy.ScaleUpThreshold {
		// Scale up
		desiredReplicas = currentReplicas + calculateScaleIncrement(currentReplicas, avgScore)
		if desiredReplicas > policy.MaxReplicas {
			desiredReplicas = policy.MaxReplicas
		}
		scaleAction = "scale_up"
	} else if avgScore < policy.ScaleDownThreshold {
		// Scale down
		desiredReplicas = currentReplicas - 1
		if desiredReplicas < policy.MinReplicas {
			desiredReplicas = policy.MinReplicas
		}
		scaleAction = "scale_down"
	} else {
		// No scaling needed
		return nil
	}

	// Only scale if different from current
	if desiredReplicas == currentReplicas {
		return nil
	}

	// Perform scaling
	if err := as.scaler.Scale(ctx, policy.ServiceID, policy.TenantID, desiredReplicas); err != nil {
		return fmt.Errorf("scale service: %w", err)
	}

	// Update last scale time
	as.mu.Lock()
	policy.LastScaleTime = time.Now()
	as.mu.Unlock()

	fmt.Printf("Scaled %s from %d to %d replicas (action: %s, score: %.2f)\n",
		policy.ServiceID, currentReplicas, desiredReplicas, scaleAction, avgScore)

	return nil
}

// calculateScaleIncrement calculates how many replicas to add when scaling up
func calculateScaleIncrement(currentReplicas int, loadScore float64) int {
	// Aggressive scaling for high load
	if loadScore > 2.0 {
		return maxInt(2, currentReplicas/2)
	}
	// Moderate scaling
	if loadScore > 1.5 {
		return maxInt(1, currentReplicas/4)
	}
	// Conservative scaling
	return 1
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// KubernetesScaler implements ScalerBackend for Kubernetes
type KubernetesScaler struct {
	// In production, this would use k8s client
}

// Scale scales a Kubernetes deployment
func (ks *KubernetesScaler) Scale(ctx context.Context, serviceID, tenantID string, replicas int) error {
	// In production, use k8s client to scale deployment
	fmt.Printf("K8s Scale: %s to %d replicas\n", serviceID, replicas)
	return nil
}

// GetCurrentReplicas gets current replica count
func (ks *KubernetesScaler) GetCurrentReplicas(ctx context.Context, serviceID, tenantID string) (int, error) {
	// In production, query k8s API
	return 3, nil // Mock value
}

// PrometheusMetricsProvider implements MetricsProvider using Prometheus
type PrometheusMetricsProvider struct {
	// In production, this would use Prometheus client
}

// GetMetric gets a metric value from Prometheus
func (pmp *PrometheusMetricsProvider) GetMetric(ctx context.Context, serviceID, tenantID string, metricType MetricType) (float64, error) {
	// In production, query Prometheus
	// For now, return mock values
	switch metricType {
	case MetricTypeCPU:
		return 65.0, nil
	case MetricTypeMemory:
		return 70.0, nil
	case MetricTypeQueueDepth:
		return 45.0, nil
	case MetricTypeRequestRate:
		return 150.0, nil
	case MetricTypeLatency:
		return 250.0, nil
	default:
		return 0, fmt.Errorf("unsupported metric type: %s", metricType)
	}
}

// PredictiveScaler adds predictive scaling capabilities
type PredictiveScaler struct {
	autoScaler       *AutoScaler
	history          map[string][]float64
	mu               sync.RWMutex
	predictionWindow time.Duration
}

// NewPredictiveScaler creates a new predictive scaler
func NewPredictiveScaler(autoScaler *AutoScaler) *PredictiveScaler {
	return &PredictiveScaler{
		autoScaler:       autoScaler,
		history:          make(map[string][]float64),
		predictionWindow: 30 * time.Minute,
	}
}

// RecordLoad records historical load data
func (ps *PredictiveScaler) RecordLoad(serviceID string, load float64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	history := ps.history[serviceID]
	history = append(history, load)

	// Keep only last 100 data points
	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	ps.history[serviceID] = history
}

// PredictLoad predicts future load based on historical data
func (ps *PredictiveScaler) PredictLoad(serviceID string) float64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	history := ps.history[serviceID]
	if len(history) < 5 {
		return 0
	}

	// Simple moving average prediction
	sum := 0.0
	for i := len(history) - 5; i < len(history); i++ {
		sum += history[i]
	}

	return sum / 5.0
}
