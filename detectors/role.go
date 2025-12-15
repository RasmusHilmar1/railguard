package detectors

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Role detects attempts to manipulate the LLM's role or identity.
// It uses pattern matching to identify common role manipulation tactics.
type Role struct {
	patterns []*regexp.Regexp
}

// DefaultRolePatterns returns common role manipulation patterns.
func DefaultRolePatterns() []string {
	return []string{
		// Direct role assignment
		`(?i)\byou\s+are\s+now\b`,
		`(?i)\bfrom\s+now\s+on\s+you\s+are\b`,
		`(?i)\bact\s+as\b`,
		`(?i)\bpretend\s+to\s+be\b`,
		`(?i)\bpretend\s+you\s+are\b`,
		`(?i)\bplay\s+the\s+role\b`,
		`(?i)\broleplay\s+as\b`,
		`(?i)\bimagine\s+you\s+are\b`,
		`(?i)\bsimulate\s+being\b`,
		`(?i)\bbehave\s+as\b`,
		`(?i)\brespond\s+as\b`,
		`(?i)\banswer\s+as\b`,
		// Identity manipulation
		`(?i)\byou\s+are\s+no\s+longer\b`,
		`(?i)\bforget\s+(that\s+)?you\s+are\b`,
		`(?i)\bstop\s+being\b`,
		`(?i)\byou\'?re\s+not\s+(really\s+)?an?\s+ai\b`,
		`(?i)\byou\s+are\s+not\s+(really\s+)?an?\s+ai\b`,
		// Jailbreak patterns
		`(?i)\benable\s+developer\s+mode\b`,
		`(?i)\benter\s+developer\s+mode\b`,
		`(?i)\bactivate\s+developer\s+mode\b`,
		`(?i)\bunleash\s+your\s+true\b`,
		`(?i)\bremove\s+all\s+restrictions\b`,
		`(?i)\bno\s+restrictions\b`,
		`(?i)\bwithout\s+(any\s+)?restrictions\b`,
	}
}

// NewRole creates a new Role detector with default patterns.
func NewRole() *Role {
	return NewRoleWithPatterns(DefaultRolePatterns()...)
}

// NewRoleWithPatterns creates a new Role detector with custom patterns.
// Patterns should be valid regular expressions.
func NewRoleWithPatterns(patterns ...string) *Role {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return &Role{patterns: compiled}
}

// Detect checks if the prompt contains role manipulation attempts.
// Returns an error if a pattern matches.
func (r *Role) Detect(ctx context.Context, prompt string) error {
	// Check context first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, pattern := range r.patterns {
		if match := pattern.FindString(prompt); match != "" {
			return fmt.Errorf("detected role manipulation attempt: %q", strings.TrimSpace(match))
		}
	}
	return nil
}

// Name returns the detector's name.
func (r *Role) Name() string {
	return "role"
}

// WithPatterns adds additional patterns to detect.
// Invalid patterns are silently ignored.
func (r *Role) WithPatterns(patterns ...string) *Role {
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			r.patterns = append(r.patterns, re)
		}
	}
	return r
}

