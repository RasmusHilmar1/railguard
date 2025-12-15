package railguard

import (
	"time"
)

// Option is a functional option for configuring a Guard.
type Option func(*Guard) error

// WithClient sets the LLM client for the Guard.
// This option is required - a Guard cannot be created without a client.
func WithClient(client Client) Option {
	return func(g *Guard) error {
		if client == nil {
			return ErrNilClient
		}
		g.client = client
		return nil
	}
}

// WithSchema sets the JSON schema for output validation.
// The provided value must be a pointer to a struct.
// When set, all LLM outputs will be validated against this schema.
func WithSchema(v interface{}) Option {
	return func(g *Guard) error {
		schema, err := NewSchema(v)
		if err != nil {
			return err
		}
		g.schema = schema
		return nil
	}
}

// WithDetectors adds one or more detectors to the Guard.
// Detectors are run in order before each generation.
// Detection failures are fatal and will not be retried.
func WithDetectors(detectors ...Detector) Option {
	return func(g *Guard) error {
		for _, d := range detectors {
			if d == nil {
				return ErrNilDetector
			}
		}
		g.detectors = append(g.detectors, detectors...)
		return nil
	}
}

// WithValidators adds one or more validators to the Guard.
// Validators are run in order after each generation.
// Validation failures may be retried based on the retry configuration.
func WithValidators(validators ...Validator) Option {
	return func(g *Guard) error {
		for _, v := range validators {
			if v == nil {
				return ErrNilValidator
			}
		}
		g.validators = append(g.validators, validators...)
		return nil
	}
}

// WithRetry sets the retry configuration for the Guard.
// This controls how generation and validation failures are retried.
func WithRetry(config RetryConfig) Option {
	return func(g *Guard) error {
		if err := config.Validate(); err != nil {
			return ErrInvalidRetryConfig
		}
		g.retry = config
		return nil
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
// This is a convenience method that modifies only the MaxAttempts field
// of the retry configuration while keeping other defaults.
func WithMaxRetries(maxRetries int) Option {
	return func(g *Guard) error {
		if maxRetries < 1 {
			return ErrInvalidRetryConfig
		}
		g.retry.MaxAttempts = maxRetries
		return nil
	}
}

// WithTimeout sets a timeout for the entire Run operation.
// The timeout applies to detection, generation, and validation combined.
// A timeout of 0 means no timeout (the context's deadline is used instead).
func WithTimeout(timeout time.Duration) Option {
	return func(g *Guard) error {
		if timeout < 0 {
			return ErrInvalidTimeout
		}
		g.timeout = timeout
		return nil
	}
}

// WithStrictSchema sets whether the schema should reject unknown fields.
// By default, strict mode is enabled to prevent hallucinated fields.
// This option only has an effect if WithSchema is also used.
func WithStrictSchema(strict bool) Option {
	return func(g *Guard) error {
		if g.schema != nil {
			g.schema.WithStrict(strict)
		}
		g.strictSchema = strict
		return nil
	}
}

