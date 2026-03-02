package cost

import (
	"fmt"
)

// OpenAIPricing implements PricingProvider for OpenAI
type OpenAIPricing struct {
	prices map[string]*ModelPricing
}

// NewOpenAIPricing creates a new OpenAI pricing provider
func NewOpenAIPricing() *OpenAIPricing {
	return &OpenAIPricing{
		prices: map[string]*ModelPricing{
			"gpt-4": {
				Provider:        "openai",
				Model:           "gpt-4",
				InputPricePerM:  30.0,
				OutputPricePerM: 60.0,
				Currency:        "USD",
			},
			"gpt-4-turbo": {
				Provider:        "openai",
				Model:           "gpt-4-turbo",
				InputPricePerM:  10.0,
				OutputPricePerM: 30.0,
				Currency:        "USD",
			},
			"gpt-3.5-turbo": {
				Provider:        "openai",
				Model:           "gpt-3.5-turbo",
				InputPricePerM:  0.5,
				OutputPricePerM: 1.5,
				Currency:        "USD",
			},
		},
	}
}

// GetModelPricing returns pricing for a specific model
func (p *OpenAIPricing) GetModelPricing(model string) (*ModelPricing, error) {
	pricing, ok := p.prices[model]
	if !ok {
		return nil, fmt.Errorf("pricing not found for model: %s", model)
	}
	return pricing, nil
}

// CalculateCost calculates the cost for given token counts
func (p *OpenAIPricing) CalculateCost(model string, inputTokens, outputTokens int64) (float64, error) {
	pricing, err := p.GetModelPricing(model)
	if err != nil {
		return 0, err
	}

	inputCost := float64(inputTokens) / 1_000_000 * pricing.InputPricePerM
	outputCost := float64(outputTokens) / 1_000_000 * pricing.OutputPricePerM
	totalCost := inputCost + outputCost

	if totalCost < pricing.MinimumCharge {
		totalCost = pricing.MinimumCharge
	}

	return totalCost, nil
}

// AnthropicPricing implements PricingProvider for Anthropic
type AnthropicPricing struct {
	prices map[string]*ModelPricing
}

// NewAnthropicPricing creates a new Anthropic pricing provider
func NewAnthropicPricing() *AnthropicPricing {
	return &AnthropicPricing{
		prices: map[string]*ModelPricing{
			"claude-3-opus": {
				Provider:        "anthropic",
				Model:           "claude-3-opus",
				InputPricePerM:  15.0,
				OutputPricePerM: 75.0,
				Currency:        "USD",
			},
			"claude-3-sonnet": {
				Provider:        "anthropic",
				Model:           "claude-3-sonnet",
				InputPricePerM:  3.0,
				OutputPricePerM: 15.0,
				Currency:        "USD",
			},
			"claude-3-haiku": {
				Provider:        "anthropic",
				Model:           "claude-3-haiku",
				InputPricePerM:  0.25,
				OutputPricePerM: 1.25,
				Currency:        "USD",
			},
		},
	}
}

// GetModelPricing returns pricing for a specific model
func (p *AnthropicPricing) GetModelPricing(model string) (*ModelPricing, error) {
	pricing, ok := p.prices[model]
	if !ok {
		return nil, fmt.Errorf("pricing not found for model: %s", model)
	}
	return pricing, nil
}

// CalculateCost calculates the cost for given token counts
func (p *AnthropicPricing) CalculateCost(model string, inputTokens, outputTokens int64) (float64, error) {
	pricing, err := p.GetModelPricing(model)
	if err != nil {
		return 0, err
	}

	inputCost := float64(inputTokens) / 1_000_000 * pricing.InputPricePerM
	outputCost := float64(outputTokens) / 1_000_000 * pricing.OutputPricePerM
	totalCost := inputCost + outputCost

	if totalCost < pricing.MinimumCharge {
		totalCost = pricing.MinimumCharge
	}

	return totalCost, nil
}

// BudgetManager manages budget limits and alerts
type BudgetManager struct {
	costTracker *CostTracker
	budgets     map[string]*Budget
}

// Budget represents a spending budget
type Budget struct {
	ID       string
	TenantID string
	AgentID  string
	UserID   string
	Limit    float64
	Period   string  // daily, weekly, monthly
	AlertAt  float64 // Percentage to trigger alert (e.g., 0.8 for 80%)
	Enabled  bool
}

// NewBudgetManager creates a new budget manager
func NewBudgetManager(tracker *CostTracker) *BudgetManager {
	return &BudgetManager{
		costTracker: tracker,
		budgets:     make(map[string]*Budget),
	}
}

// SetBudget sets a budget limit
func (bm *BudgetManager) SetBudget(budget *Budget) {
	bm.budgets[budget.ID] = budget
}

// CheckBudget checks if a budget is exceeded
func (bm *BudgetManager) CheckBudget(budgetID string) (bool, float64, error) {
	_, ok := bm.budgets[budgetID]
	if !ok {
		return false, 0, fmt.Errorf("budget not found: %s", budgetID)
	}

	// Calculate time range based on period
	// (simplified - real implementation would use proper period calculation)
	// For now, return false (not exceeded)
	return false, 0, nil
}
