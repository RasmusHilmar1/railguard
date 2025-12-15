// Package validators provides built-in validator implementations for railguard.
package validators

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// JSON validates that the output is valid JSON.
// It can optionally check for specific JSON structures (object, array, etc.).
type JSON struct {
	requireObject bool
	requireArray  bool
}

// NewJSON creates a new JSON validator that accepts any valid JSON.
func NewJSON() *JSON {
	return &JSON{}
}

// RequireObject returns a new JSON validator that requires the output to be a JSON object.
func (j *JSON) RequireObject() *JSON {
	return &JSON{
		requireObject: true,
		requireArray:  false,
	}
}

// RequireArray returns a new JSON validator that requires the output to be a JSON array.
func (j *JSON) RequireArray() *JSON {
	return &JSON{
		requireObject: false,
		requireArray:  true,
	}
}

// Validate checks if the output is valid JSON.
// Returns an error if the output is not valid JSON or doesn't match the required structure.
func (j *JSON) Validate(ctx context.Context, output string) error {
	// Check context first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Trim whitespace for structure detection
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return fmt.Errorf("empty output is not valid JSON")
	}

	// Quick check for structure requirements before full parse
	if j.requireObject && !strings.HasPrefix(trimmed, "{") {
		return fmt.Errorf("expected JSON object, got output starting with %q", firstChar(trimmed))
	}
	if j.requireArray && !strings.HasPrefix(trimmed, "[") {
		return fmt.Errorf("expected JSON array, got output starting with %q", firstChar(trimmed))
	}

	// Validate JSON syntax
	var v interface{}
	if err := json.Unmarshal([]byte(output), &v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Additional structure validation
	if j.requireObject {
		if _, ok := v.(map[string]interface{}); !ok {
			return fmt.Errorf("expected JSON object, got %T", v)
		}
	}
	if j.requireArray {
		if _, ok := v.([]interface{}); !ok {
			return fmt.Errorf("expected JSON array, got %T", v)
		}
	}

	return nil
}

// Name returns the validator's name.
func (j *JSON) Name() string {
	return "json"
}

// firstChar returns the first character of a string for error messages.
func firstChar(s string) string {
	if len(s) == 0 {
		return ""
	}
	for _, r := range s {
		return string(r)
	}
	return ""
}

// JSONExtractor extracts JSON from text that may contain markdown code blocks.
// This is useful when LLMs wrap JSON in ```json ... ``` blocks.
type JSONExtractor struct {
	inner *JSON
}

// NewJSONExtractor creates a new JSONExtractor with the given JSON validator.
func NewJSONExtractor() *JSONExtractor {
	return &JSONExtractor{inner: NewJSON()}
}

// WithValidator sets the inner JSON validator to use after extraction.
func (e *JSONExtractor) WithValidator(v *JSON) *JSONExtractor {
	e.inner = v
	return e
}

// Extract attempts to extract JSON from the output.
// It handles markdown code blocks (```json ... ```) and bare JSON.
func (e *JSONExtractor) Extract(output string) string {
	trimmed := strings.TrimSpace(output)

	// Try to extract from markdown code block
	if strings.HasPrefix(trimmed, "```") {
		// Find the end of the first line (language specifier)
		firstNewline := strings.Index(trimmed, "\n")
		if firstNewline == -1 {
			return trimmed
		}

		// Find the closing ```
		rest := trimmed[firstNewline+1:]
		closingIndex := strings.LastIndex(rest, "```")
		if closingIndex == -1 {
			return trimmed
		}

		return strings.TrimSpace(rest[:closingIndex])
	}

	return trimmed
}

// Validate extracts JSON from the output and validates it.
func (e *JSONExtractor) Validate(ctx context.Context, output string) error {
	extracted := e.Extract(output)
	return e.inner.Validate(ctx, extracted)
}

// Name returns the validator's name.
func (e *JSONExtractor) Name() string {
	return "json_extractor"
}

