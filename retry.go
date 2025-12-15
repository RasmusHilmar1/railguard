package railguard

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryConfig configures the retry behavior for generation and validation.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (including the initial attempt).
	// Must be at least 1.
	MaxAttempts int

	// InitialDelay is the delay before the first retry.
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry.
	Multiplier float64

	// Jitter adds randomness to delays to prevent thundering herd.
	// Value between 0 and 1, where 0.1 means ±10% jitter.
	Jitter float64
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
	}
}

// Validate checks that the retry configuration is valid.
func (c RetryConfig) Validate() error {
	if c.MaxAttempts < 1 {
		return errors.New("max attempts must be at least 1")
	}
	if c.InitialDelay < 0 {
		return errors.New("initial delay cannot be negative")
	}
	if c.MaxDelay < 0 {
		return errors.New("max delay cannot be negative")
	}
	if c.Multiplier < 1 {
		return errors.New("multiplier must be at least 1")
	}
	if c.Jitter < 0 || c.Jitter > 1 {
		return errors.New("jitter must be between 0 and 1")
	}
	return nil
}

// delay calculates the delay for a given attempt number (0-indexed).
func (c RetryConfig) delay(attempt int) time.Duration {
	if attempt == 0 || c.InitialDelay == 0 {
		return 0
	}

	// Calculate exponential backoff
	delay := float64(c.InitialDelay) * math.Pow(c.Multiplier, float64(attempt-1))

	// Cap at max delay
	if delay > float64(c.MaxDelay) {
		delay = float64(c.MaxDelay)
	}

	// Apply jitter
	if c.Jitter > 0 {
		jitterRange := delay * c.Jitter
		delay = delay - jitterRange + (rand.Float64() * 2 * jitterRange)
	}

	return time.Duration(delay)
}

// backoff sleeps for the appropriate delay for the given attempt.
// Returns immediately if the context is canceled.
func (c RetryConfig) backoff(ctx context.Context, attempt int) error {
	delay := c.delay(attempt)
	if delay == 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// retryClassifier determines whether an error is retryable.
type retryClassifier struct{}

// fatalErrors are errors that should not be retried.
// These typically indicate permanent failures or security issues.
var fatalErrors = []error{
	context.Canceled,
	context.DeadlineExceeded,
}

// shouldRetry determines if an error is retryable.
// Detection errors and context errors are never retried.
// Generation and validation errors are generally retryable.
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Detection errors are never retried - they represent security failures
	var detectionErr *DetectionError
	if errors.As(err, &detectionErr) {
		return false
	}

	// Context errors are never retried
	for _, fatalErr := range fatalErrors {
		if errors.Is(err, fatalErr) {
			return false
		}
	}

	// Generation and validation errors are generally retryable
	// as LLM outputs are non-deterministic
	return true
}

