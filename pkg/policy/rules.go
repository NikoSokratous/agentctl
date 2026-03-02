package policy

// Action is the policy decision.
type Action string

const (
	ActionAllow           Action = "allow"
	ActionDeny            Action = "deny"
	ActionRequireApproval Action = "require_approval"
)

// Rule is a single policy rule.
type Rule struct {
	Name      string    `yaml:"name"`
	Match     MatchSpec `yaml:"match"`
	Action    Action    `yaml:"action"`
	Message   string    `yaml:"message"`
	Approvers []string  `yaml:"approvers"`
}

// MatchSpec describes when a rule applies.
type MatchSpec struct {
	Tool        string `yaml:"tool"`
	Condition   string `yaml:"condition"` // CEL expression
	Environment string `yaml:"environment"`
	RiskScore   string `yaml:"risk_score"` // CEL expression, e.g. ">= 0.8"
}

// PolicyConfig is the schema for policy.yaml.
type PolicyConfig struct {
	Version string `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}
