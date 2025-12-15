package railguard_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RasmusHilmar1/railguard"
)

type mockClient struct{}

func (m *mockClient) Generate(ctx context.Context, prompt string) (string, error) {
	return "response", nil
}

type mockDetector struct {
	name string
}

func (m *mockDetector) Detect(ctx context.Context, prompt string) error { return nil }
func (m *mockDetector) Name() string                                    { return m.name }

type mockValidator struct {
	name string
}

func (m *mockValidator) Validate(ctx context.Context, output string) error { return nil }
func (m *mockValidator) Name() string                                      { return m.name }

func TestWithClient(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		g, err := railguard.New(railguard.WithClient(&mockClient{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.Client() == nil {
			t.Error("client should not be nil")
		}
	})

	t.Run("nil client", func(t *testing.T) {
		_, err := railguard.New(railguard.WithClient(nil))
		if !errors.Is(err, railguard.ErrNilClient) {
			t.Errorf("expected ErrNilClient, got %v", err)
		}
	})
}

func TestWithSchema(t *testing.T) {
	type Response struct {
		Data string `json:"data"`
	}

	t.Run("valid schema", func(t *testing.T) {
		g, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithSchema(&Response{}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.Schema() == nil {
			t.Error("schema should not be nil")
		}
	})

	t.Run("invalid schema", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithSchema("not a struct"),
		)
		if !errors.Is(err, railguard.ErrInvalidSchema) {
			t.Errorf("expected ErrInvalidSchema, got %v", err)
		}
	})
}

func TestWithDetectors(t *testing.T) {
	t.Run("valid detectors", func(t *testing.T) {
		d1 := &mockDetector{name: "d1"}
		d2 := &mockDetector{name: "d2"}

		g, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithDetectors(d1, d2),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		detectors := g.Detectors()
		if len(detectors) != 2 {
			t.Errorf("expected 2 detectors, got %d", len(detectors))
		}
	})

	t.Run("nil detector", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithDetectors(nil),
		)
		if !errors.Is(err, railguard.ErrNilDetector) {
			t.Errorf("expected ErrNilDetector, got %v", err)
		}
	})

	t.Run("multiple calls accumulate", func(t *testing.T) {
		d1 := &mockDetector{name: "d1"}
		d2 := &mockDetector{name: "d2"}

		g, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithDetectors(d1),
			railguard.WithDetectors(d2),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(g.Detectors()) != 2 {
			t.Errorf("expected 2 detectors, got %d", len(g.Detectors()))
		}
	})
}

func TestWithValidators(t *testing.T) {
	t.Run("valid validators", func(t *testing.T) {
		v1 := &mockValidator{name: "v1"}
		v2 := &mockValidator{name: "v2"}

		g, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithValidators(v1, v2),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		validators := g.Validators()
		if len(validators) != 2 {
			t.Errorf("expected 2 validators, got %d", len(validators))
		}
	})

	t.Run("nil validator", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithValidators(nil),
		)
		if !errors.Is(err, railguard.ErrNilValidator) {
			t.Errorf("expected ErrNilValidator, got %v", err)
		}
	})
}

func TestWithRetry(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := railguard.RetryConfig{
			MaxAttempts:  5,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     10 * time.Second,
			Multiplier:   3.0,
			Jitter:       0.2,
		}

		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithRetry(config),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		config := railguard.RetryConfig{
			MaxAttempts: 0, // Invalid
		}

		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithRetry(config),
		)
		if !errors.Is(err, railguard.ErrInvalidRetryConfig) {
			t.Errorf("expected ErrInvalidRetryConfig, got %v", err)
		}
	})
}

func TestWithMaxRetries(t *testing.T) {
	t.Run("valid max retries", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithMaxRetries(5),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid max retries", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithMaxRetries(0),
		)
		if !errors.Is(err, railguard.ErrInvalidRetryConfig) {
			t.Errorf("expected ErrInvalidRetryConfig, got %v", err)
		}
	})
}

func TestWithTimeout(t *testing.T) {
	t.Run("valid timeout", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithTimeout(5*time.Second),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("zero timeout allowed", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithTimeout(0),
		)
		if err != nil {
			t.Fatalf("zero timeout should be allowed: %v", err)
		}
	})

	t.Run("negative timeout", func(t *testing.T) {
		_, err := railguard.New(
			railguard.WithClient(&mockClient{}),
			railguard.WithTimeout(-1*time.Second),
		)
		if !errors.Is(err, railguard.ErrInvalidTimeout) {
			t.Errorf("expected ErrInvalidTimeout, got %v", err)
		}
	})
}

