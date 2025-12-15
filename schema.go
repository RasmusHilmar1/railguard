package railguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// Schema enforces JSON structure matching a Go type.
// It validates that LLM output conforms to an expected structure,
// preventing hallucinated fields and ensuring type safety.
type Schema struct {
	targetType reflect.Type
	strict     bool
}

// NewSchema creates a Schema from a struct pointer.
// The provided value must be a pointer to a struct; other types will return an error.
//
// Example:
//
//	type Response struct {
//	    Result string `json:"result"`
//	}
//	schema, err := railguard.NewSchema(&Response{})
func NewSchema(v interface{}) (*Schema, error) {
	if v == nil {
		return nil, ErrInvalidSchema
	}

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("%w: got %v, want pointer", ErrInvalidSchema, t.Kind())
	}

	elem := t.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: got pointer to %v, want struct", ErrInvalidSchema, elem.Kind())
	}

	return &Schema{
		targetType: elem,
		strict:     true, // Default to strict to prevent hallucinated fields
	}, nil
}

// WithStrict sets whether the schema should reject unknown fields.
// By default, strict mode is enabled to prevent hallucinated fields.
func (s *Schema) WithStrict(strict bool) *Schema {
	s.strict = strict
	return s
}

// IsStrict returns whether the schema is in strict mode.
func (s *Schema) IsStrict() bool {
	return s.strict
}

// TargetType returns the reflect.Type that this schema validates against.
func (s *Schema) TargetType() reflect.Type {
	return s.targetType
}

// Unmarshal parses JSON data into a new instance of the schema's target type.
// In strict mode (default), unknown fields in the JSON will cause an error.
// Returns a pointer to the populated struct or an error if parsing fails.
func (s *Schema) Unmarshal(data []byte) (interface{}, error) {
	// Create a new instance of the target type
	ptr := reflect.New(s.targetType)

	// Set up decoder with optional strict mode
	dec := json.NewDecoder(bytes.NewReader(data))
	if s.strict {
		dec.DisallowUnknownFields()
	}

	// Decode the JSON
	if err := dec.Decode(ptr.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Check for trailing data after the JSON value
	if dec.More() {
		return nil, fmt.Errorf("trailing data after JSON")
	}

	return ptr.Interface(), nil
}

// UnmarshalInto parses JSON data into the provided destination.
// The destination must be a pointer to a struct of the same type as the schema.
// In strict mode (default), unknown fields in the JSON will cause an error.
func (s *Schema) UnmarshalInto(data []byte, dest interface{}) error {
	if dest == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer, got %v", destType.Kind())
	}

	if destType.Elem() != s.targetType {
		return fmt.Errorf("destination type mismatch: got %v, want %v", destType.Elem(), s.targetType)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if s.strict {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if dec.More() {
		return fmt.Errorf("trailing data after JSON")
	}

	return nil
}

// Validate checks if the provided JSON data can be unmarshaled into the schema's target type.
// Returns nil if valid, or an error describing why validation failed.
func (s *Schema) Validate(data []byte) error {
	_, err := s.Unmarshal(data)
	return err
}

