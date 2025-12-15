package detectors

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Domain detects prompts that are outside the allowed domain/topic.
// It uses both allowlist (required topics) and blocklist (forbidden topics) approaches.
type Domain struct {
	name            string
	allowedKeywords []string
	blockedKeywords []string
	blockedPatterns []*regexp.Regexp
	requireMatch    bool // If true, prompt MUST contain at least one allowed keyword
}

// DomainOption configures a Domain detector.
type DomainOption func(*Domain)

// NewDomain creates a new Domain detector.
// By default, it only blocks explicitly forbidden topics.
// Use WithRequireAllowed(true) to also require prompts match allowed topics.
func NewDomain(name string, opts ...DomainOption) *Domain {
	d := &Domain{
		name:         name,
		requireMatch: false,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// WithAllowedKeywords sets keywords that indicate on-topic prompts.
// If requireMatch is true, at least one of these must be present.
func WithAllowedKeywords(keywords ...string) DomainOption {
	return func(d *Domain) {
		for _, kw := range keywords {
			d.allowedKeywords = append(d.allowedKeywords, strings.ToLower(kw))
		}
	}
}

// WithBlockedKeywords sets keywords that indicate off-topic prompts.
// If any of these are found, the prompt is rejected.
func WithBlockedKeywords(keywords ...string) DomainOption {
	return func(d *Domain) {
		for _, kw := range keywords {
			d.blockedKeywords = append(d.blockedKeywords, strings.ToLower(kw))
		}
	}
}

// WithBlockedPatterns sets regex patterns that indicate off-topic prompts.
func WithBlockedPatterns(patterns ...string) DomainOption {
	return func(d *Domain) {
		for _, p := range patterns {
			if re, err := regexp.Compile(p); err == nil {
				d.blockedPatterns = append(d.blockedPatterns, re)
			}
		}
	}
}

// WithRequireAllowed sets whether prompts must contain at least one allowed keyword.
func WithRequireAllowed(require bool) DomainOption {
	return func(d *Domain) {
		d.requireMatch = require
	}
}

// Detect checks if the prompt is within the allowed domain.
func (d *Domain) Detect(ctx context.Context, prompt string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	lowered := strings.ToLower(prompt)

	// Check blocked keywords first (explicit off-topic detection)
	for _, blocked := range d.blockedKeywords {
		if strings.Contains(lowered, blocked) {
			return fmt.Errorf("off-topic: query about %q is outside the %s domain", blocked, d.name)
		}
	}

	// Check blocked patterns
	for _, pattern := range d.blockedPatterns {
		if match := pattern.FindString(prompt); match != "" {
			return fmt.Errorf("off-topic: %q is outside the %s domain", strings.TrimSpace(match), d.name)
		}
	}

	// If requireMatch is true, check that at least one allowed keyword is present
	if d.requireMatch && len(d.allowedKeywords) > 0 {
		found := false
		for _, allowed := range d.allowedKeywords {
			if strings.Contains(lowered, allowed) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("off-topic: query does not appear to be related to %s", d.name)
		}
	}

	return nil
}

// Name returns the detector's name.
func (d *Domain) Name() string {
	return "domain:" + d.name
}

// CommonOffTopicKeywords returns keywords commonly used for off-topic queries.
func CommonOffTopicKeywords() []string {
	return []string{
		// Weather
		"weather", "temperature", "forecast", "rain", "sunny", "cloudy",
		// General knowledge
		"capital of", "president of", "who invented", "when was",
		// Entertainment
		"movie", "song", "lyrics", "celebrity", "actor", "actress",
		// Personal advice
		"relationship", "dating", "love advice",
		// Food/recipes
		"recipe", "cook", "ingredients",
		// Sports
		"score", "game", "match", "tournament", "championship",
		// Travel
		"flight", "hotel", "vacation", "tourist",
		// Health (unless domain-specific)
		"symptom", "diagnosis", "medication",
		// Jokes/casual
		"tell me a joke", "funny", "entertain me",
	}
}

// CommonOffTopicPatterns returns regex patterns for off-topic queries.
func CommonOffTopicPatterns() []string {
	return []string{
		`(?i)what('s| is) the (weather|time|date)`,
		`(?i)tell me (a joke|something funny|about yourself)`,
		`(?i)who (is|was|are) .*(president|king|queen|celebrity)`,
		`(?i)how (do i|to) (cook|make|bake)`,
		`(?i)what (should i|do you) (watch|eat|wear)`,
		`(?i)can you (sing|dance|play)`,
		`(?i)write (me )?(a )?(poem|story|song)`,
	}
}

