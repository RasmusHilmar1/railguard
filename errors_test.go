package railguard_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/RasmusHilmar1/railguard"
)

func TestDetectionError(t *testing.T) {
	innerErr := errors.New("suspicious content")
	err := &railguard.DetectionError{
		Detector: "keywords",
		Err:      innerErr,
	}

	t.Run("error message", func(t *testing.T) {
		msg := err.Error()
		if !strings.Contains(msg, "detection failed") {
			t.Errorf("expected 'detection failed' in message, got %q", msg)
		}
		if !strings.Contains(msg, "keywords") {
			t.Errorf("expected detector name in message, got %q", msg)
		}
		if !strings.Contains(msg, "suspicious content") {
			t.Errorf("expected inner error in message, got %q", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		if unwrapped := err.Unwrap(); unwrapped != innerErr {
			t.Errorf("expected unwrapped error to be %v, got %v", innerErr, unwrapped)
		}
	})

	t.Run("errors.Is", func(t *testing.T) {
		if !errors.Is(err, innerErr) {
			t.Error("errors.Is should find inner error")
		}
	})
}

func TestValidationError(t *testing.T) {
	innerErr := errors.New("invalid JSON")
	err := &railguard.ValidationError{
		Validator: "json",
		Err:       innerErr,
	}

	t.Run("error message", func(t *testing.T) {
		msg := err.Error()
		if !strings.Contains(msg, "validation failed") {
			t.Errorf("expected 'validation failed' in message, got %q", msg)
		}
		if !strings.Contains(msg, "json") {
			t.Errorf("expected validator name in message, got %q", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		if unwrapped := err.Unwrap(); unwrapped != innerErr {
			t.Errorf("expected unwrapped error to be %v, got %v", innerErr, unwrapped)
		}
	})
}

func TestSchemaError(t *testing.T) {
	innerErr := errors.New("missing field")
	err := &railguard.SchemaError{
		Err: innerErr,
	}

	t.Run("error message", func(t *testing.T) {
		msg := err.Error()
		if !strings.Contains(msg, "schema validation failed") {
			t.Errorf("expected 'schema validation failed' in message, got %q", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		if unwrapped := err.Unwrap(); unwrapped != innerErr {
			t.Errorf("expected unwrapped error to be %v, got %v", innerErr, unwrapped)
		}
	})
}

func TestGenerationError(t *testing.T) {
	innerErr := errors.New("network timeout")
	err := &railguard.GenerationError{
		Err: innerErr,
	}

	t.Run("error message", func(t *testing.T) {
		msg := err.Error()
		if !strings.Contains(msg, "generation failed") {
			t.Errorf("expected 'generation failed' in message, got %q", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		if unwrapped := err.Unwrap(); unwrapped != innerErr {
			t.Errorf("expected unwrapped error to be %v, got %v", innerErr, unwrapped)
		}
	})
}

func TestMaxRetriesError(t *testing.T) {
	innerErr := errors.New("last attempt failed")
	err := &railguard.MaxRetriesError{
		Attempts: 3,
		LastErr:  innerErr,
	}

	t.Run("error message", func(t *testing.T) {
		msg := err.Error()
		if !strings.Contains(msg, "max retries exceeded") {
			t.Errorf("expected 'max retries exceeded' in message, got %q", msg)
		}
		if !strings.Contains(msg, "3") {
			t.Errorf("expected attempt count in message, got %q", msg)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		if unwrapped := err.Unwrap(); unwrapped != innerErr {
			t.Errorf("expected unwrapped error to be %v, got %v", innerErr, unwrapped)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		railguard.ErrNoClient,
		railguard.ErrNoSchema,
		railguard.ErrNilClient,
		railguard.ErrNilDetector,
		railguard.ErrNilValidator,
		railguard.ErrInvalidRetryConfig,
		railguard.ErrInvalidTimeout,
		railguard.ErrInvalidSchema,
	}

	for _, sentinel := range sentinels {
		t.Run(sentinel.Error(), func(t *testing.T) {
			if sentinel == nil {
				t.Error("sentinel error should not be nil")
			}
			if sentinel.Error() == "" {
				t.Error("sentinel error message should not be empty")
			}
			if !strings.HasPrefix(sentinel.Error(), "railguard:") {
				t.Errorf("sentinel error should have 'railguard:' prefix, got %q", sentinel.Error())
			}
		})
	}
}

