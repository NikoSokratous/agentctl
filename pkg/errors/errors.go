package errors

import (
	"errors"
	"fmt"
)

// Error codes for structured error handling
const (
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodePolicyDenied     = "POLICY_DENIED"
	ErrCodeToolNotFound     = "TOOL_NOT_FOUND"
	ErrCodeToolExecutionErr = "TOOL_EXECUTION_ERROR"
	ErrCodeLLMError         = "LLM_ERROR"
	ErrCodeStoreError       = "STORE_ERROR"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeCancelled        = "CANCELLED"
	ErrCodeInternal         = "INTERNAL_ERROR"
)

// AgentError is a structured error with code and context.
type AgentError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]any
}

// Error implements the error interface.
func (e *AgentError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *AgentError) Unwrap() error {
	return e.Cause
}

// New creates a new AgentError.
func New(code, message string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Context: make(map[string]any),
	}
}

// Wrap wraps an existing error with code and message.
func Wrap(err error, code, message string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Cause:   err,
		Context: make(map[string]any),
	}
}

// WithContext adds context to the error.
func (e *AgentError) WithContext(key string, value any) *AgentError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// Is checks if an error is of a specific code.
func Is(err error, code string) bool {
	var ae *AgentError
	if errors.As(err, &ae) {
		return ae.Code == code
	}
	return false
}

// Convenience constructors

func NotFound(message string) *AgentError {
	return New(ErrCodeNotFound, message)
}

func InvalidInput(message string) *AgentError {
	return New(ErrCodeInvalidInput, message)
}

func Unauthorized(message string) *AgentError {
	return New(ErrCodeUnauthorized, message)
}

func Forbidden(message string) *AgentError {
	return New(ErrCodeForbidden, message)
}

func PolicyDenied(message string) *AgentError {
	return New(ErrCodePolicyDenied, message)
}

func ToolNotFound(toolName string) *AgentError {
	return New(ErrCodeToolNotFound, fmt.Sprintf("tool not found: %s", toolName)).
		WithContext("tool", toolName)
}

func ToolExecutionError(toolName string, err error) *AgentError {
	return Wrap(err, ErrCodeToolExecutionErr, fmt.Sprintf("tool execution failed: %s", toolName)).
		WithContext("tool", toolName)
}

func LLMError(provider string, err error) *AgentError {
	return Wrap(err, ErrCodeLLMError, fmt.Sprintf("LLM error: %s", provider)).
		WithContext("provider", provider)
}

func StoreError(operation string, err error) *AgentError {
	return Wrap(err, ErrCodeStoreError, fmt.Sprintf("store error: %s", operation)).
		WithContext("operation", operation)
}

func Timeout(message string) *AgentError {
	return New(ErrCodeTimeout, message)
}

func Cancelled(message string) *AgentError {
	return New(ErrCodeCancelled, message)
}

func Internal(message string, err error) *AgentError {
	return Wrap(err, ErrCodeInternal, message)
}
