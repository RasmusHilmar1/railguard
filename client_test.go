package railguard_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RasmusHilmar1/railguard"
)

func TestClientFunc(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			return "Hello, " + prompt, nil
		})

		result, err := client.Generate(context.Background(), "World")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Hello, World" {
			t.Errorf("expected 'Hello, World', got %q", result)
		}
	})

	t.Run("returns error", func(t *testing.T) {
		expectedErr := errors.New("generation failed")
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			return "", expectedErr
		})

		_, err := client.Generate(context.Background(), "test")
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "success", nil
			}
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.Generate(ctx, "test")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

