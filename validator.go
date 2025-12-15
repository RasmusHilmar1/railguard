package railguard

import "context"

// Validator checks LLM outputs after generation.
// Validators are used for output validation like JSON syntax checking,
// length limits, content filtering, or any post-generation validation.
//
// Validation failures may be retried depending on the Guard's configuration,
// as LLM outputs are non-deterministic.
type Validator interface {
	// Validate examines the output and returns an error if it should be rejected.
	// A nil return indicates the output passed validation.
	Validate(ctx context.Context, output string) error

	// Name returns a human-readable identifier for this validator.
	// Used in error messages and logging.
	Name() string
}

// ValidatorFunc is an adapter that allows ordinary functions to be used as Validators.
// The Name() method returns "custom" for function-based validators.
type ValidatorFunc func(ctx context.Context, output string) error

// Validate implements the Validator interface by calling the function itself.
func (f ValidatorFunc) Validate(ctx context.Context, output string) error {
	return f(ctx, output)
}

// Name returns "custom" for function-based validators.
func (f ValidatorFunc) Name() string {
	return "custom"
}

// Validators is a convenience type for working with multiple validators.
type Validators []Validator

// Validate runs all validators in sequence, returning the first error encountered.
func (v Validators) Validate(ctx context.Context, output string) error {
	for _, validator := range v {
		if err := validator.Validate(ctx, output); err != nil {
			return err
		}
	}
	return nil
}

