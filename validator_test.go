package railguard_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RasmusHilmar1/railguard"
)

func TestValidatorFunc(t *testing.T) {
	t.Run("successful validation", func(t *testing.T) {
		validator := railguard.ValidatorFunc(func(ctx context.Context, output string) error {
			return nil
		})

		if err := validator.Validate(context.Background(), "test"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error", func(t *testing.T) {
		expectedErr := errors.New("validation failed")
		validator := railguard.ValidatorFunc(func(ctx context.Context, output string) error {
			return expectedErr
		})

		err := validator.Validate(context.Background(), "test")
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("has correct name", func(t *testing.T) {
		validator := railguard.ValidatorFunc(func(ctx context.Context, output string) error {
			return nil
		})

		if name := validator.Name(); name != "custom" {
			t.Errorf("expected name 'custom', got %q", name)
		}
	})
}

func TestValidators(t *testing.T) {
	t.Run("empty validators pass", func(t *testing.T) {
		validators := railguard.Validators{}
		if err := validators.Validate(context.Background(), "test"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("all pass", func(t *testing.T) {
		validators := railguard.Validators{
			railguard.ValidatorFunc(func(ctx context.Context, output string) error { return nil }),
			railguard.ValidatorFunc(func(ctx context.Context, output string) error { return nil }),
		}
		if err := validators.Validate(context.Background(), "test"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("stops on first error", func(t *testing.T) {
		expectedErr := errors.New("first error")
		callCount := 0

		validators := railguard.Validators{
			railguard.ValidatorFunc(func(ctx context.Context, output string) error {
				callCount++
				return expectedErr
			}),
			railguard.ValidatorFunc(func(ctx context.Context, output string) error {
				callCount++
				return errors.New("second error")
			}),
		}

		err := validators.Validate(context.Background(), "test")
		if err != expectedErr {
			t.Errorf("expected first error, got %v", err)
		}
		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})
}

