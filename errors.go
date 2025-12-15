package railguard

import (
	"errors"
	"fmt"
)

// Configuration errors - sentinel errors for common configuration mistakes.
var (
	// ErrNoClient is returned when Guard is created without a client.
	ErrNoClient = errors.New("railguard: no client provided")

	// ErrNoSchema is returned when schema validation is required but no schema was provided.
	ErrNoSchema = errors.New("railguard: no schema provided")

	// ErrNilClient is returned when a nil client is passed to WithClient.
	ErrNilClient = errors.New("railguard: client cannot be nil")

	// ErrNilDetector is returned when a nil detector is passed to WithDetectors.
	ErrNilDetector = errors.New("railguard: detector cannot be nil")

	// ErrNilValidator is returned when a nil validator is passed to WithValidators.
	ErrNilValidator = errors.New("railguard: validator cannot be nil")

	// ErrInvalidRetryConfig is returned when retry configuration is invalid.
	ErrInvalidRetryConfig = errors.New("railguard: invalid retry configuration")

	// ErrInvalidTimeout is returned when a non-positive timeout is provided.
	ErrInvalidTimeout = errors.New("railguard: timeout must be positive")

	// ErrInvalidSchema is returned when the schema is not a pointer to a struct.
	ErrInvalidSchema = errors.New("railguard: schema must be a pointer to a struct")
)

// DetectionError wraps errors from detectors with context about which detector failed.
type DetectionError struct {
	// Detector is the name of the detector that failed.
	Detector string
	// Err is the underlying error from the detector.
	Err error
}

// Error implements the error interface.
func (e *DetectionError) Error() string {
	return fmt.Sprintf("detection failed [%s]: %v", e.Detector, e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *DetectionError) Unwrap() error {
	return e.Err
}

// ValidationError wraps errors from validators with context about which validator failed.
type ValidationError struct {
	// Validator is the name of the validator that failed.
	Validator string
	// Err is the underlying error from the validator.
	Err error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed [%s]: %v", e.Validator, e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// SchemaError wraps errors from schema validation.
type SchemaError struct {
	// Err is the underlying JSON unmarshaling or validation error.
	Err error
}

// Error implements the error interface.
func (e *SchemaError) Error() string {
	return fmt.Sprintf("schema validation failed: %v", e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *SchemaError) Unwrap() error {
	return e.Err
}

// GenerationError wraps errors from the LLM client.
type GenerationError struct {
	// Err is the underlying error from the client.
	Err error
}

// Error implements the error interface.
func (e *GenerationError) Error() string {
	return fmt.Sprintf("generation failed: %v", e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *GenerationError) Unwrap() error {
	return e.Err
}

// MaxRetriesError is returned when the maximum number of retry attempts is exceeded.
type MaxRetriesError struct {
	// Attempts is the number of attempts made.
	Attempts int
	// LastErr is the error from the final attempt.
	LastErr error
}

// Error implements the error interface.
func (e *MaxRetriesError) Error() string {
	return fmt.Sprintf("max retries exceeded after %d attempts: %v", e.Attempts, e.LastErr)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *MaxRetriesError) Unwrap() error {
	return e.LastErr
}

