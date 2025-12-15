package validators_test

import (
	"context"
	"testing"

	"github.com/RasmusHilmar1/railguard/validators"
)

func TestMaxLength(t *testing.T) {
	t.Run("within limit", func(t *testing.T) {
		v := validators.NewMaxLength(100)
		err := v.Validate(context.Background(), "short text")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("at limit", func(t *testing.T) {
		v := validators.NewMaxLength(5)
		err := v.Validate(context.Background(), "12345")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("exceeds limit", func(t *testing.T) {
		v := validators.NewMaxLength(5)
		err := v.Validate(context.Background(), "123456")
		if err == nil {
			t.Error("expected error for exceeding limit")
		}
	})

	t.Run("counts bytes by default", func(t *testing.T) {
		v := validators.NewMaxLength(5)
		// "hello" = 5 bytes, should pass
		err := v.Validate(context.Background(), "hello")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// "こんにちは" = 15 bytes (3 bytes per character), should fail
		err = v.Validate(context.Background(), "こんにちは")
		if err == nil {
			t.Error("expected error for multi-byte string exceeding byte limit")
		}
	})

	t.Run("WithRunes counts characters", func(t *testing.T) {
		v := validators.NewMaxLength(5).WithRunes()

		// "hello" = 5 characters, should pass
		err := v.Validate(context.Background(), "hello")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// "こんにちは" = 5 characters, should pass
		err = v.Validate(context.Background(), "こんにちは")
		if err != nil {
			t.Errorf("unexpected error for 5-character string: %v", err)
		}

		// "こんにちは!" = 6 characters, should fail
		err = v.Validate(context.Background(), "こんにちは!")
		if err == nil {
			t.Error("expected error for 6-character string")
		}
	})

	t.Run("Name returns correct value", func(t *testing.T) {
		v := validators.NewMaxLength(100)
		if v.Name() != "maxlength" {
			t.Errorf("expected name 'maxlength', got %q", v.Name())
		}
	})

	t.Run("Limit accessor", func(t *testing.T) {
		v := validators.NewMaxLength(42)
		if v.Limit() != 42 {
			t.Errorf("expected limit 42, got %d", v.Limit())
		}
	})

	t.Run("CountsRunes accessor", func(t *testing.T) {
		v := validators.NewMaxLength(100)
		if v.CountsRunes() {
			t.Error("expected CountsRunes to be false by default")
		}

		v = v.WithRunes()
		if !v.CountsRunes() {
			t.Error("expected CountsRunes to be true after WithRunes")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		v := validators.NewMaxLength(100)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := v.Validate(ctx, "test")
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestMinLength(t *testing.T) {
	t.Run("meets minimum", func(t *testing.T) {
		v := validators.NewMinLength(5)
		err := v.Validate(context.Background(), "hello world")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("at minimum", func(t *testing.T) {
		v := validators.NewMinLength(5)
		err := v.Validate(context.Background(), "12345")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("below minimum", func(t *testing.T) {
		v := validators.NewMinLength(5)
		err := v.Validate(context.Background(), "1234")
		if err == nil {
			t.Error("expected error for below minimum")
		}
	})

	t.Run("Name returns correct value", func(t *testing.T) {
		v := validators.NewMinLength(5)
		if v.Name() != "minlength" {
			t.Errorf("expected name 'minlength', got %q", v.Name())
		}
	})
}

func TestLengthRange(t *testing.T) {
	t.Run("within range", func(t *testing.T) {
		v := validators.NewLengthRange(5, 10)
		err := v.Validate(context.Background(), "1234567")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("at minimum", func(t *testing.T) {
		v := validators.NewLengthRange(5, 10)
		err := v.Validate(context.Background(), "12345")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("at maximum", func(t *testing.T) {
		v := validators.NewLengthRange(5, 10)
		err := v.Validate(context.Background(), "1234567890")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("below minimum", func(t *testing.T) {
		v := validators.NewLengthRange(5, 10)
		err := v.Validate(context.Background(), "1234")
		if err == nil {
			t.Error("expected error for below minimum")
		}
	})

	t.Run("exceeds maximum", func(t *testing.T) {
		v := validators.NewLengthRange(5, 10)
		err := v.Validate(context.Background(), "12345678901")
		if err == nil {
			t.Error("expected error for exceeds maximum")
		}
	})

	t.Run("Name returns correct value", func(t *testing.T) {
		v := validators.NewLengthRange(5, 10)
		if v.Name() != "lengthrange" {
			t.Errorf("expected name 'lengthrange', got %q", v.Name())
		}
	})
}

