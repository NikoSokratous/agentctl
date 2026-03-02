package tool

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// InputSchema validates tool input. Uses JSON Schema.
type InputSchema struct {
	schema *jsonschema.Schema
	raw    []byte
}

// NewInputSchema creates an InputSchema from a JSON Schema document.
func NewInputSchema(raw []byte) (*InputSchema, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(raw)); err != nil {
		return nil, err
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, err
	}
	return &InputSchema{schema: schema, raw: raw}, nil
}

// Validate validates input against the schema.
func (s *InputSchema) Validate(input []byte) error {
	var v any
	if err := json.Unmarshal(input, &v); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return s.schema.Validate(v)
}
