package railguard

import (
	"context"
	"time"
)

// Guard orchestrates LLM interactions with safety checks and validation.
// It provides a pipeline for:
//  1. Detection - Pre-generation safety checks (e.g., prompt injection detection)
//  2. Generation - Calling the LLM client
//  3. Validation - Post-generation output validation (e.g., JSON syntax)
//  4. Schema - Structured output parsing
//
// Detection failures are fatal and not retried. Generation and validation
// failures may be retried based on the retry configuration.
type Guard struct {
	client       Client
	detectors    []Detector
	validators   []Validator
	schema       *Schema
	retry        RetryConfig
	timeout      time.Duration
	strictSchema bool
}

// Result contains the output from a successful Guard.Run call.
type Result struct {
	// Raw is the raw string output from the LLM.
	Raw string

	// Parsed is the structured output, populated when a schema is configured.
	// It will be a pointer to the schema's target type.
	Parsed interface{}

	// Metadata contains information about the execution.
	Metadata Metadata
}

// Metadata contains information about a Guard.Run execution.
type Metadata struct {
	// Attempts is the number of generation attempts made (1 = success on first try).
	Attempts int

	// Duration is the total time spent in Run, including all retries.
	Duration time.Duration
}

// New creates a new Guard with the provided options.
// At minimum, a client must be provided via WithClient.
//
// Example:
//
//	guard, err := railguard.New(
//	    railguard.WithClient(myClient),
//	    railguard.WithSchema(&Response{}),
//	    railguard.WithMaxRetries(3),
//	)
func New(opts ...Option) (*Guard, error) {
	g := &Guard{
		retry: DefaultRetryConfig(),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}

	// Validate required fields
	if g.client == nil {
		return nil, ErrNoClient
	}

	// Apply strict schema setting if schema was created before the option
	if g.schema != nil && !g.strictSchema {
		// strictSchema defaults to false, but schema defaults to strict=true
		// Only change if explicitly set to non-strict
	}

	return g, nil
}

// Run executes the Guard pipeline for the given prompt.
// The pipeline consists of:
//  1. Apply timeout (if configured)
//  2. Run detectors (fail fast, no retry)
//  3. Retry loop: generate → validate → parse schema
//
// Returns a Result on success, or an error if the pipeline fails.
func (g *Guard) Run(ctx context.Context, prompt string) (*Result, error) {
	startTime := time.Now()

	// Apply timeout if configured
	if g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	// Phase 1: Detection (fail fast, no retry)
	if err := g.runDetectors(ctx, prompt); err != nil {
		return nil, err
	}

	// Phase 2 & 3: Generation, Validation, and Schema (with retry)
	var lastErr error
	for attempt := 0; attempt < g.retry.MaxAttempts; attempt++ {
		// Backoff before retry (not before first attempt)
		if attempt > 0 {
			if err := g.retry.backoff(ctx, attempt); err != nil {
				return nil, err
			}
		}

		// Generate
		output, err := g.client.Generate(ctx, prompt)
		if err != nil {
			lastErr = &GenerationError{Err: err}
			if !shouldRetry(lastErr) {
				return nil, lastErr
			}
			continue
		}

		// Validate
		if err := g.runValidators(ctx, output); err != nil {
			lastErr = err
			if !shouldRetry(lastErr) {
				return nil, lastErr
			}
			continue
		}

		// Parse schema
		parsed, err := g.parseSchema(output)
		if err != nil {
			lastErr = &SchemaError{Err: err}
			if !shouldRetry(lastErr) {
				return nil, lastErr
			}
			continue
		}

		// Success!
		return &Result{
			Raw:    output,
			Parsed: parsed,
			Metadata: Metadata{
				Attempts: attempt + 1,
				Duration: time.Since(startTime),
			},
		}, nil
	}

	// Max retries exceeded
	return nil, &MaxRetriesError{
		Attempts: g.retry.MaxAttempts,
		LastErr:  lastErr,
	}
}

// runDetectors runs all detectors in sequence.
// Returns a DetectionError on the first failure.
func (g *Guard) runDetectors(ctx context.Context, prompt string) error {
	for _, detector := range g.detectors {
		if err := detector.Detect(ctx, prompt); err != nil {
			return &DetectionError{
				Detector: detector.Name(),
				Err:      err,
			}
		}
	}
	return nil
}

// runValidators runs all validators in sequence.
// Returns a ValidationError on the first failure.
func (g *Guard) runValidators(ctx context.Context, output string) error {
	for _, validator := range g.validators {
		if err := validator.Validate(ctx, output); err != nil {
			return &ValidationError{
				Validator: validator.Name(),
				Err:       err,
			}
		}
	}
	return nil
}

// parseSchema parses the output using the configured schema.
// Returns nil, nil if no schema is configured.
func (g *Guard) parseSchema(output string) (interface{}, error) {
	if g.schema == nil {
		return nil, nil
	}
	return g.schema.Unmarshal([]byte(output))
}

// Client returns the configured client.
func (g *Guard) Client() Client {
	return g.client
}

// Schema returns the configured schema, or nil if not set.
func (g *Guard) Schema() *Schema {
	return g.schema
}

// Detectors returns a copy of the configured detectors.
func (g *Guard) Detectors() []Detector {
	if g.detectors == nil {
		return nil
	}
	result := make([]Detector, len(g.detectors))
	copy(result, g.detectors)
	return result
}

// Validators returns a copy of the configured validators.
func (g *Guard) Validators() []Validator {
	if g.validators == nil {
		return nil
	}
	result := make([]Validator, len(g.validators))
	copy(result, g.validators)
	return result
}

