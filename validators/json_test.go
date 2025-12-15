package validators_test

import (
	"context"
	"testing"

	"github.com/RasmusHilmar1/railguard/validators"
)

func TestJSON(t *testing.T) {
	t.Run("valid JSON object", func(t *testing.T) {
		v := validators.NewJSON()
		err := v.Validate(context.Background(), `{"key": "value"}`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		v := validators.NewJSON()
		err := v.Validate(context.Background(), `[1, 2, 3]`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid JSON primitive", func(t *testing.T) {
		v := validators.NewJSON()

		primitives := []string{`"string"`, `123`, `true`, `false`, `null`}
		for _, p := range primitives {
			err := v.Validate(context.Background(), p)
			if err != nil {
				t.Errorf("unexpected error for %q: %v", p, err)
			}
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		v := validators.NewJSON()

		invalid := []string{
			`{invalid}`,
			`{"key": }`,
			`[1, 2, ]`,
			`not json`,
			``,
			`   `,
		}

		for _, i := range invalid {
			err := v.Validate(context.Background(), i)
			if err == nil {
				t.Errorf("expected error for %q", i)
			}
		}
	})

	t.Run("RequireObject", func(t *testing.T) {
		v := validators.NewJSON().RequireObject()

		// Object should pass
		err := v.Validate(context.Background(), `{"key": "value"}`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Array should fail
		err = v.Validate(context.Background(), `[1, 2, 3]`)
		if err == nil {
			t.Error("expected error for array when RequireObject")
		}

		// Primitive should fail
		err = v.Validate(context.Background(), `"string"`)
		if err == nil {
			t.Error("expected error for primitive when RequireObject")
		}
	})

	t.Run("RequireArray", func(t *testing.T) {
		v := validators.NewJSON().RequireArray()

		// Array should pass
		err := v.Validate(context.Background(), `[1, 2, 3]`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Object should fail
		err = v.Validate(context.Background(), `{"key": "value"}`)
		if err == nil {
			t.Error("expected error for object when RequireArray")
		}
	})

	t.Run("Name returns correct value", func(t *testing.T) {
		v := validators.NewJSON()
		if v.Name() != "json" {
			t.Errorf("expected name 'json', got %q", v.Name())
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		v := validators.NewJSON()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := v.Validate(ctx, `{"key": "value"}`)
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestJSONExtractor(t *testing.T) {
	t.Run("extracts from markdown code block", func(t *testing.T) {
		e := validators.NewJSONExtractor()

		output := "```json\n{\"key\": \"value\"}\n```"
		err := e.Validate(context.Background(), output)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("extracts from bare JSON", func(t *testing.T) {
		e := validators.NewJSONExtractor()

		output := `{"key": "value"}`
		err := e.Validate(context.Background(), output)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("handles whitespace", func(t *testing.T) {
		e := validators.NewJSONExtractor()

		output := "  \n  {\"key\": \"value\"}  \n  "
		err := e.Validate(context.Background(), output)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Extract method", func(t *testing.T) {
		e := validators.NewJSONExtractor()

		tests := []struct {
			input    string
			expected string
		}{
			{`{"key": "value"}`, `{"key": "value"}`},
			{"```json\n{\"key\": \"value\"}\n```", `{"key": "value"}`},
			{"  {\"key\": \"value\"}  ", `{"key": "value"}`},
		}

		for _, tt := range tests {
			result := e.Extract(tt.input)
			if result != tt.expected {
				t.Errorf("Extract(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Name returns correct value", func(t *testing.T) {
		e := validators.NewJSONExtractor()
		if e.Name() != "json_extractor" {
			t.Errorf("expected name 'json_extractor', got %q", e.Name())
		}
	})
}

