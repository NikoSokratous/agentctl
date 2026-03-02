package workflow

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

// CELEvaluator evaluates CEL expressions for workflow conditions.
type CELEvaluator struct {
	env *cel.Env
}

// NewCELEvaluator creates a new CEL evaluator.
func NewCELEvaluator() (*CELEvaluator, error) {
	// Create CEL environment with workflow-specific variables
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("outputs", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewVar("step", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewVar("workflow", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create CEL environment: %w", err)
	}

	return &CELEvaluator{env: env}, nil
}

// Evaluate evaluates a CEL condition with the given context.
func (e *CELEvaluator) Evaluate(condition string, context map[string]interface{}) (bool, error) {
	// Parse the expression
	ast, issues := e.env.Compile(condition)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("compile condition: %w", issues.Err())
	}

	// Create program
	prg, err := e.env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("create program: %w", err)
	}

	// Evaluate
	result, _, err := prg.Eval(context)
	if err != nil {
		return false, fmt.Errorf("evaluate condition: %w", err)
	}

	// Convert to bool
	boolResult, ok := result.Value().(bool)
	if !ok {
		return false, fmt.Errorf("condition must evaluate to boolean, got %T", result.Value())
	}

	return boolResult, nil
}

// ValidateCondition validates a CEL expression without evaluating it.
func (e *CELEvaluator) ValidateCondition(condition string) error {
	_, issues := e.env.Compile(condition)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("invalid condition: %w", issues.Err())
	}

	return nil
}

// GetAvailableVariables returns the variables available in conditions.
func (e *CELEvaluator) GetAvailableVariables() []string {
	return []string{
		"outputs",           // map[string]interface{} - Previous step outputs
		"step.status",       // string - Current step status
		"step.error",        // string - Error message if failed
		"workflow.duration", // duration - Workflow elapsed time
	}
}

// SimpleCELEvaluator is a simple CEL evaluator without advanced features.
type SimpleCELEvaluator struct{}

// NewSimpleCELEvaluator creates a basic CEL evaluator.
func NewSimpleCELEvaluator() *SimpleCELEvaluator {
	return &SimpleCELEvaluator{}
}

// Evaluate evaluates simple conditions.
func (s *SimpleCELEvaluator) Evaluate(condition string, context map[string]interface{}) (bool, error) {
	// Create a simple CEL environment
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("outputs", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return false, err
	}

	ast, issues := env.Compile(condition)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, err
	}

	result, _, err := prg.Eval(context)
	if err != nil {
		return false, err
	}

	if boolVal, ok := result.Value().(bool); ok {
		return boolVal, nil
	}

	return false, fmt.Errorf("condition did not evaluate to boolean")
}
