package validators

import (
	"context"
	"fmt"
	"unicode/utf8"
)

// MaxLength validates that the output does not exceed a specified length.
// It can count either bytes or runes (characters).
type MaxLength struct {
	limit int
	runes bool // if true, counts runes (characters) instead of bytes
}

// NewMaxLength creates a new MaxLength validator that counts bytes.
func NewMaxLength(limit int) *MaxLength {
	return &MaxLength{
		limit: limit,
		runes: false,
	}
}

// WithRunes configures the validator to count runes (characters) instead of bytes.
// This is useful for internationalized text where byte count != character count.
func (m *MaxLength) WithRunes() *MaxLength {
	return &MaxLength{
		limit: m.limit,
		runes: true,
	}
}

// Validate checks if the output length is within the limit.
func (m *MaxLength) Validate(ctx context.Context, output string) error {
	// Check context first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var length int
	var unit string

	if m.runes {
		length = utf8.RuneCountInString(output)
		unit = "characters"
	} else {
		length = len(output)
		unit = "bytes"
	}

	if length > m.limit {
		return fmt.Errorf("output length %d %s exceeds limit of %d", length, unit, m.limit)
	}

	return nil
}

// Name returns the validator's name.
func (m *MaxLength) Name() string {
	return "maxlength"
}

// Limit returns the configured limit.
func (m *MaxLength) Limit() int {
	return m.limit
}

// CountsRunes returns true if the validator counts runes instead of bytes.
func (m *MaxLength) CountsRunes() bool {
	return m.runes
}

// MinLength validates that the output is at least a specified length.
// It can count either bytes or runes (characters).
type MinLength struct {
	limit int
	runes bool
}

// NewMinLength creates a new MinLength validator that counts bytes.
func NewMinLength(limit int) *MinLength {
	return &MinLength{
		limit: limit,
		runes: false,
	}
}

// WithRunes configures the validator to count runes (characters) instead of bytes.
func (m *MinLength) WithRunes() *MinLength {
	return &MinLength{
		limit: m.limit,
		runes: true,
	}
}

// Validate checks if the output length meets the minimum requirement.
func (m *MinLength) Validate(ctx context.Context, output string) error {
	// Check context first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var length int
	var unit string

	if m.runes {
		length = utf8.RuneCountInString(output)
		unit = "characters"
	} else {
		length = len(output)
		unit = "bytes"
	}

	if length < m.limit {
		return fmt.Errorf("output length %d %s is below minimum of %d", length, unit, m.limit)
	}

	return nil
}

// Name returns the validator's name.
func (m *MinLength) Name() string {
	return "minlength"
}

// LengthRange validates that the output length is within a specified range.
type LengthRange struct {
	min   int
	max   int
	runes bool
}

// NewLengthRange creates a new LengthRange validator that counts bytes.
func NewLengthRange(min, max int) *LengthRange {
	return &LengthRange{
		min:   min,
		max:   max,
		runes: false,
	}
}

// WithRunes configures the validator to count runes (characters) instead of bytes.
func (r *LengthRange) WithRunes() *LengthRange {
	return &LengthRange{
		min:   r.min,
		max:   r.max,
		runes: true,
	}
}

// Validate checks if the output length is within the specified range.
func (r *LengthRange) Validate(ctx context.Context, output string) error {
	// Check context first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var length int
	var unit string

	if r.runes {
		length = utf8.RuneCountInString(output)
		unit = "characters"
	} else {
		length = len(output)
		unit = "bytes"
	}

	if length < r.min {
		return fmt.Errorf("output length %d %s is below minimum of %d", length, unit, r.min)
	}
	if length > r.max {
		return fmt.Errorf("output length %d %s exceeds maximum of %d", length, unit, r.max)
	}

	return nil
}

// Name returns the validator's name.
func (r *LengthRange) Name() string {
	return "lengthrange"
}

