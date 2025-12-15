package railguard

import "context"

// Detector checks prompts before they are sent to the LLM.
// Detectors are used for safety checks like prompt injection detection,
// content filtering, or any pre-generation validation.
//
// Detection failures are considered fatal and will not be retried.
// This is by design: if a prompt is unsafe, retrying won't make it safe.
type Detector interface {
	// Detect examines the prompt and returns an error if it should be rejected.
	// A nil return indicates the prompt passed detection.
	Detect(ctx context.Context, prompt string) error

	// Name returns a human-readable identifier for this detector.
	// Used in error messages and logging.
	Name() string
}

// DetectorFunc is an adapter that allows ordinary functions to be used as Detectors.
// The Name() method returns "custom" for function-based detectors.
type DetectorFunc func(ctx context.Context, prompt string) error

// Detect implements the Detector interface by calling the function itself.
func (f DetectorFunc) Detect(ctx context.Context, prompt string) error {
	return f(ctx, prompt)
}

// Name returns "custom" for function-based detectors.
func (f DetectorFunc) Name() string {
	return "custom"
}

// Detectors is a convenience type for working with multiple detectors.
type Detectors []Detector

// Detect runs all detectors in sequence, returning the first error encountered.
func (d Detectors) Detect(ctx context.Context, prompt string) error {
	for _, detector := range d {
		if err := detector.Detect(ctx, prompt); err != nil {
			return err
		}
	}
	return nil
}

