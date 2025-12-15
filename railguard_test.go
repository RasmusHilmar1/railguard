package railguard_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/RasmusHilmar1/railguard"
)

func TestNew(t *testing.T) {
	t.Run("no client", func(t *testing.T) {
		_, err := railguard.New()
		if !errors.Is(err, railguard.ErrNoClient) {
			t.Errorf("expected ErrNoClient, got %v", err)
		}
	})

	t.Run("with client only", func(t *testing.T) {
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			return "response", nil
		})

		g, err := railguard.New(railguard.WithClient(client))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g == nil {
			t.Fatal("guard should not be nil")
		}
	})
}

func TestGuardRun(t *testing.T) {
	t.Run("simple generation", func(t *testing.T) {
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			return "Hello, " + prompt, nil
		})

		g, err := railguard.New(railguard.WithClient(client))
		if err != nil {
			t.Fatalf("failed to create guard: %v", err)
		}

		result, err := g.Run(context.Background(), "World")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Raw != "Hello, World" {
			t.Errorf("expected 'Hello, World', got %q", result.Raw)
		}
		if result.Metadata.Attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", result.Metadata.Attempts)
		}
	})

	t.Run("with schema parsing", func(t *testing.T) {
		type Response struct {
			Message string `json:"message"`
		}

		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			return `{"message": "success"}`, nil
		})

		g, err := railguard.New(
			railguard.WithClient(client),
			railguard.WithSchema(&Response{}),
		)
		if err != nil {
			t.Fatalf("failed to create guard: %v", err)
		}

		result, err := g.Run(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp, ok := result.Parsed.(*Response)
		if !ok {
			t.Fatalf("expected *Response, got %T", result.Parsed)
		}
		if resp.Message != "success" {
			t.Errorf("expected message 'success', got %q", resp.Message)
		}
	})

	t.Run("detector blocks prompt", func(t *testing.T) {
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			return "should not reach here", nil
		})

		detector := railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
			if strings.Contains(prompt, "blocked") {
				return errors.New("prompt blocked")
			}
			return nil
		})

		g, err := railguard.New(
			railguard.WithClient(client),
			railguard.WithDetectors(detector),
		)
		if err != nil {
			t.Fatalf("failed to create guard: %v", err)
		}

		_, err = g.Run(context.Background(), "this is blocked")
		if err == nil {
			t.Fatal("expected error for blocked prompt")
		}

		var detectionErr *railguard.DetectionError
		if !errors.As(err, &detectionErr) {
			t.Errorf("expected DetectionError, got %T", err)
		}
	})

	t.Run("validator rejects output", func(t *testing.T) {
		attempts := 0
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			attempts++
			return "invalid output", nil
		})

		validator := railguard.ValidatorFunc(func(ctx context.Context, output string) error {
			return errors.New("validation failed")
		})

		g, err := railguard.New(
			railguard.WithClient(client),
			railguard.WithValidators(validator),
			railguard.WithMaxRetries(2),
		)
		if err != nil {
			t.Fatalf("failed to create guard: %v", err)
		}

		_, err = g.Run(context.Background(), "test")
		if err == nil {
			t.Fatal("expected error for invalid output")
		}

		var maxRetriesErr *railguard.MaxRetriesError
		if !errors.As(err, &maxRetriesErr) {
			t.Errorf("expected MaxRetriesError, got %T: %v", err, err)
		}

		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("retry on generation failure", func(t *testing.T) {
		attempts := 0
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			attempts++
			if attempts < 3 {
				return "", errors.New("temporary failure")
			}
			return "success", nil
		})

		g, err := railguard.New(
			railguard.WithClient(client),
			railguard.WithMaxRetries(3),
		)
		if err != nil {
			t.Fatalf("failed to create guard: %v", err)
		}

		result, err := g.Run(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Raw != "success" {
			t.Errorf("expected 'success', got %q", result.Raw)
		}
		if result.Metadata.Attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", result.Metadata.Attempts)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(1 * time.Second):
				return "too slow", nil
			}
		})

		g, err := railguard.New(
			railguard.WithClient(client),
			railguard.WithTimeout(50*time.Millisecond),
		)
		if err != nil {
			t.Fatalf("failed to create guard: %v", err)
		}

		_, err = g.Run(context.Background(), "test")
		if err == nil {
			t.Fatal("expected timeout error")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		})

		g, err := railguard.New(railguard.WithClient(client))
		if err != nil {
			t.Fatalf("failed to create guard: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err = g.Run(ctx, "test")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestGuardAccessors(t *testing.T) {
	client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
		return "response", nil
	})

	t.Run("returns nil for empty detectors", func(t *testing.T) {
		g, _ := railguard.New(railguard.WithClient(client))
		if g.Detectors() != nil {
			t.Error("expected nil detectors")
		}
	})

	t.Run("returns nil for empty validators", func(t *testing.T) {
		g, _ := railguard.New(railguard.WithClient(client))
		if g.Validators() != nil {
			t.Error("expected nil validators")
		}
	})

	t.Run("returns nil for no schema", func(t *testing.T) {
		g, _ := railguard.New(railguard.WithClient(client))
		if g.Schema() != nil {
			t.Error("expected nil schema")
		}
	})
}

