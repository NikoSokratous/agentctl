package policy

import "context"

// ApprovalCallback is called when human approval is required.
type ApprovalCallback func(ctx context.Context, tool string, input map[string]any, approvers []string) (approved bool, err error)

// ApprovalGate blocks until approval is granted or denied.
type ApprovalGate struct {
	Callback ApprovalCallback
}

// NewApprovalGate creates an approval gate.
func NewApprovalGate(cb ApprovalCallback) *ApprovalGate {
	return &ApprovalGate{Callback: cb}
}

// RequestApproval invokes the callback and returns the decision.
func (g *ApprovalGate) RequestApproval(ctx context.Context, tool string, input map[string]any, approvers []string) (bool, error) {
	if g.Callback == nil {
		return false, nil
	}
	return g.Callback(ctx, tool, input, approvers)
}
