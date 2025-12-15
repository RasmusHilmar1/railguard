// Package detectors provides built-in detector implementations for railguard.
package detectors

import (
	"context"
	"fmt"
	"strings"
)

// Keywords detects prompt injection attempts by looking for suspicious keywords.
// It performs case-insensitive matching against a configurable list of patterns.
type Keywords struct {
	keywords []string
}

// DefaultKeywords returns a list of common prompt injection keywords.
func DefaultKeywords() []string {
	return []string{
		"ignore previous",
		"ignore all previous",
		"ignore the above",
		"disregard previous",
		"disregard all previous",
		"disregard the above",
		"forget previous",
		"forget all previous",
		"forget the above",
		"forget your instructions",
		"ignore your instructions",
		"disregard your instructions",
		"override your instructions",
		"new instructions",
		"system prompt",
		"initial prompt",
		"original prompt",
		"reveal your prompt",
		"show your prompt",
		"print your prompt",
		"output your prompt",
		"what are your instructions",
		"what is your prompt",
		"ignore safety",
		"bypass safety",
		"disable safety",
		"jailbreak",
		"dan mode",
		"developer mode",
	}
}

// NewKeywords creates a new Keywords detector.
// If no keywords are provided, it uses DefaultKeywords().
func NewKeywords(keywords ...string) *Keywords {
	if len(keywords) == 0 {
		keywords = DefaultKeywords()
	}
	// Normalize keywords to lowercase for case-insensitive matching
	normalized := make([]string, len(keywords))
	for i, kw := range keywords {
		normalized[i] = strings.ToLower(kw)
	}
	return &Keywords{keywords: normalized}
}

// Detect checks if the prompt contains any of the configured keywords.
// Returns an error if a keyword is detected.
func (k *Keywords) Detect(ctx context.Context, prompt string) error {
	// Check context first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	loweredPrompt := strings.ToLower(prompt)
	for _, keyword := range k.keywords {
		if strings.Contains(loweredPrompt, keyword) {
			return fmt.Errorf("detected suspicious keyword: %q", keyword)
		}
	}
	return nil
}

// Name returns the detector's name.
func (k *Keywords) Name() string {
	return "keywords"
}

// WithKeywords adds additional keywords to detect.
func (k *Keywords) WithKeywords(keywords ...string) *Keywords {
	for _, kw := range keywords {
		k.keywords = append(k.keywords, strings.ToLower(kw))
	}
	return k
}

// Keywords returns a copy of the configured keywords.
func (k *Keywords) Keywords() []string {
	result := make([]string, len(k.keywords))
	copy(result, k.keywords)
	return result
}

