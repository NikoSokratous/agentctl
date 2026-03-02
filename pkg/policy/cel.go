package policy

import (
	"context"
	"fmt"

	cel "github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// EvalContext provides variables for CEL evaluation.
type EvalContext struct {
	Tool        string
	Input       map[string]any
	Environment string
	RiskScore   float64
}

// Eval evaluates a CEL expression with the given context.
func Eval(expr string, ctx EvalContext) (ref.Val, error) {
	if expr == "" {
		return types.Bool(true), nil
	}

	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("tool", decls.String),
			decls.NewVar("input", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewVar("environment", decls.String),
			decls.NewVar("risk_score", decls.Double),
			decls.NewVar("input.to", decls.String),
		),
	)
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	inputMap := types.DefaultTypeAdapter.NativeToValue(ctx.Input)
	if inputMap == nil {
		inputMap = types.DefaultTypeAdapter.NativeToValue(map[string]any{})
	}

	// Provide input.to if input has "to" field
	var inputTo string
	if t, ok := ctx.Input["to"].(string); ok {
		inputTo = t
	}

	out, _, err := prg.ContextEval(context.Background(), map[string]any{
		"tool":        ctx.Tool,
		"input":       inputMap,
		"environment": ctx.Environment,
		"risk_score":  ctx.RiskScore,
		"input.to":    inputTo,
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EvalBool evaluates a CEL expression and returns a boolean.
func EvalBool(expr string, ctx EvalContext) (bool, error) {
	v, err := Eval(expr, ctx)
	if err != nil {
		return false, err
	}
	b, ok := v.(types.Bool)
	if !ok {
		return false, fmt.Errorf("expected bool, got %T", v)
	}
	return bool(b), nil
}
