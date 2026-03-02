package builtin

import (
	"github.com/agentruntime/agentruntime/pkg/tool"
)

// All returns all built-in tools.
func All() []tool.Tool {
	return []tool.Tool{
		&HTTPRequest{},
		&Echo{},
		&Calc{},
	}
}
