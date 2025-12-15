package railguard_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/RasmusHilmar1/railguard"
)

type testResponse struct {
	Result  string `json:"result"`
	Count   int    `json:"count"`
	Success bool   `json:"success"`
}

func TestNewSchema(t *testing.T) {
	t.Run("valid struct pointer", func(t *testing.T) {
		schema, err := railguard.NewSchema(&testResponse{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if schema == nil {
			t.Fatal("schema should not be nil")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		_, err := railguard.NewSchema(nil)
		if !errors.Is(err, railguard.ErrInvalidSchema) {
			t.Errorf("expected ErrInvalidSchema, got %v", err)
		}
	})

	t.Run("non-pointer", func(t *testing.T) {
		_, err := railguard.NewSchema(testResponse{})
		if !errors.Is(err, railguard.ErrInvalidSchema) {
			t.Errorf("expected ErrInvalidSchema, got %v", err)
		}
	})

	t.Run("pointer to non-struct", func(t *testing.T) {
		str := "hello"
		_, err := railguard.NewSchema(&str)
		if !errors.Is(err, railguard.ErrInvalidSchema) {
			t.Errorf("expected ErrInvalidSchema, got %v", err)
		}
	})
}

func TestSchemaUnmarshal(t *testing.T) {
	schema, err := railguard.NewSchema(&testResponse{})
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Run("valid JSON", func(t *testing.T) {
		data := []byte(`{"result": "success", "count": 42, "success": true}`)
		result, err := schema.Unmarshal(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp, ok := result.(*testResponse)
		if !ok {
			t.Fatalf("expected *testResponse, got %T", result)
		}

		if resp.Result != "success" {
			t.Errorf("expected result 'success', got %q", resp.Result)
		}
		if resp.Count != 42 {
			t.Errorf("expected count 42, got %d", resp.Count)
		}
		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		data := []byte(`{invalid}`)
		_, err := schema.Unmarshal(data)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("unknown fields rejected in strict mode", func(t *testing.T) {
		data := []byte(`{"result": "test", "unknown_field": "value"}`)
		_, err := schema.Unmarshal(data)
		if err == nil {
			t.Error("expected error for unknown fields in strict mode")
		}
	})

	t.Run("trailing data rejected", func(t *testing.T) {
		data := []byte(`{"result": "test"}extra data`)
		_, err := schema.Unmarshal(data)
		if err == nil {
			t.Error("expected error for trailing data")
		}
	})
}

func TestSchemaStrictMode(t *testing.T) {
	schema, err := railguard.NewSchema(&testResponse{})
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Run("default is strict", func(t *testing.T) {
		if !schema.IsStrict() {
			t.Error("schema should be strict by default")
		}
	})

	t.Run("can disable strict mode", func(t *testing.T) {
		schema.WithStrict(false)
		if schema.IsStrict() {
			t.Error("schema should not be strict after WithStrict(false)")
		}

		// Now unknown fields should be allowed
		data := []byte(`{"result": "test", "unknown_field": "value"}`)
		_, err := schema.Unmarshal(data)
		if err != nil {
			t.Errorf("non-strict mode should allow unknown fields: %v", err)
		}
	})
}

func TestSchemaUnmarshalInto(t *testing.T) {
	schema, err := railguard.NewSchema(&testResponse{})
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Run("valid destination", func(t *testing.T) {
		data := []byte(`{"result": "test", "count": 1}`)
		var dest testResponse
		err := schema.UnmarshalInto(data, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest.Result != "test" {
			t.Errorf("expected result 'test', got %q", dest.Result)
		}
	})

	t.Run("nil destination", func(t *testing.T) {
		data := []byte(`{"result": "test"}`)
		err := schema.UnmarshalInto(data, nil)
		if err == nil {
			t.Error("expected error for nil destination")
		}
	})

	t.Run("wrong type destination", func(t *testing.T) {
		type otherResponse struct {
			Data string `json:"data"`
		}
		data := []byte(`{"result": "test"}`)
		var dest otherResponse
		err := schema.UnmarshalInto(data, &dest)
		if err == nil {
			t.Error("expected error for wrong type destination")
		}
		if !strings.Contains(err.Error(), "type mismatch") {
			t.Errorf("expected type mismatch error, got %v", err)
		}
	})
}

func TestSchemaValidate(t *testing.T) {
	schema, err := railguard.NewSchema(&testResponse{})
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Run("valid JSON passes", func(t *testing.T) {
		data := []byte(`{"result": "test"}`)
		if err := schema.Validate(data); err != nil {
			t.Errorf("expected validation to pass: %v", err)
		}
	})

	t.Run("invalid JSON fails", func(t *testing.T) {
		data := []byte(`{not valid}`)
		if err := schema.Validate(data); err == nil {
			t.Error("expected validation to fail for invalid JSON")
		}
	})
}

