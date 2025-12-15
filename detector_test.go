package railguard_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RasmusHilmar1/railguard"
)

func TestDetectorFunc(t *testing.T) {
	t.Run("successful detection", func(t *testing.T) {
		detector := railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
			return nil
		})

		if err := detector.Detect(context.Background(), "test"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error", func(t *testing.T) {
		expectedErr := errors.New("prompt rejected")
		detector := railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
			return expectedErr
		})

		err := detector.Detect(context.Background(), "test")
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("has correct name", func(t *testing.T) {
		detector := railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
			return nil
		})

		if name := detector.Name(); name != "custom" {
			t.Errorf("expected name 'custom', got %q", name)
		}
	})
}

func TestDetectors(t *testing.T) {
	t.Run("empty detectors pass", func(t *testing.T) {
		detectors := railguard.Detectors{}
		if err := detectors.Detect(context.Background(), "test"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("all pass", func(t *testing.T) {
		detectors := railguard.Detectors{
			railguard.DetectorFunc(func(ctx context.Context, prompt string) error { return nil }),
			railguard.DetectorFunc(func(ctx context.Context, prompt string) error { return nil }),
		}
		if err := detectors.Detect(context.Background(), "test"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("stops on first error", func(t *testing.T) {
		expectedErr := errors.New("first error")
		callCount := 0

		detectors := railguard.Detectors{
			railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
				callCount++
				return expectedErr
			}),
			railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
				callCount++
				return errors.New("second error")
			}),
		}

		err := detectors.Detect(context.Background(), "test")
		if err != expectedErr {
			t.Errorf("expected first error, got %v", err)
		}
		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})
}

