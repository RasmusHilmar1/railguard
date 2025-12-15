package detectors_test

import (
	"context"
	"strings"
	"testing"

	"github.com/RasmusHilmar1/railguard/detectors"
)

func TestDomain(t *testing.T) {
	t.Run("blocks off-topic keywords", func(t *testing.T) {
		d := detectors.NewDomain("invoice",
			detectors.WithBlockedKeywords("weather", "recipe", "movie"),
		)

		tests := []struct {
			prompt    string
			shouldErr bool
		}{
			{"Find invoice #12345", false},
			{"What's the weather today?", true},
			{"Give me a recipe for pasta", true},
			{"What movies are playing?", true},
			{"Search invoices from last month", false},
		}

		for _, tt := range tests {
			err := d.Detect(context.Background(), tt.prompt)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for %q", tt.prompt)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.prompt, err)
			}
		}
	})

	t.Run("requires allowed keywords when enabled", func(t *testing.T) {
		d := detectors.NewDomain("invoice",
			detectors.WithAllowedKeywords("invoice", "payment", "billing", "receipt"),
			detectors.WithRequireAllowed(true),
		)

		tests := []struct {
			prompt    string
			shouldErr bool
		}{
			{"Find invoice #12345", false},
			{"Show me my payments", false},
			{"Check billing status", false},
			{"What is 2+2?", true},                 // No allowed keyword
			{"Hello, how are you?", true},         // No allowed keyword
			{"Tell me about the company", true},   // No allowed keyword
		}

		for _, tt := range tests {
			err := d.Detect(context.Background(), tt.prompt)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for %q", tt.prompt)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.prompt, err)
			}
		}
	})

	t.Run("blocks patterns", func(t *testing.T) {
		d := detectors.NewDomain("invoice",
			detectors.WithBlockedPatterns(
				`(?i)what('s| is) the weather`,
				`(?i)tell me a joke`,
			),
		)

		tests := []struct {
			prompt    string
			shouldErr bool
		}{
			{"Find invoice for client ABC", false},
			{"What's the weather like?", true},
			{"Tell me a joke please", true},
			{"What is the status of invoice 123?", false},
		}

		for _, tt := range tests {
			err := d.Detect(context.Background(), tt.prompt)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for %q", tt.prompt)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.prompt, err)
			}
		}
	})

	t.Run("combined allowlist and blocklist", func(t *testing.T) {
		d := detectors.NewDomain("invoice",
			detectors.WithAllowedKeywords("invoice", "payment", "billing"),
			detectors.WithBlockedKeywords("weather", "joke"),
			detectors.WithRequireAllowed(true),
		)

		tests := []struct {
			prompt    string
			shouldErr bool
		}{
			{"Find invoice #123", false},
			{"Weather for invoice delivery", true}, // Blocked wins
			{"What's the weather?", true},
			{"Random question", true}, // No allowed keyword
		}

		for _, tt := range tests {
			err := d.Detect(context.Background(), tt.prompt)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for %q", tt.prompt)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.prompt, err)
			}
		}
	})

	t.Run("name includes domain", func(t *testing.T) {
		d := detectors.NewDomain("invoice")
		if !strings.Contains(d.Name(), "invoice") {
			t.Errorf("name should contain domain: %s", d.Name())
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		d := detectors.NewDomain("invoice")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := d.Detect(ctx, "test")
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestCommonOffTopicKeywords(t *testing.T) {
	keywords := detectors.CommonOffTopicKeywords()
	if len(keywords) == 0 {
		t.Error("should return some keywords")
	}
}

func TestCommonOffTopicPatterns(t *testing.T) {
	patterns := detectors.CommonOffTopicPatterns()
	if len(patterns) == 0 {
		t.Error("should return some patterns")
	}

	// Verify patterns compile
	d := detectors.NewDomain("test",
		detectors.WithBlockedPatterns(patterns...),
	)

	// Test that common off-topic queries are blocked
	offTopicQueries := []string{
		"What's the weather today?",
		"Tell me a joke",
		"Who is the president?",
	}

	for _, q := range offTopicQueries {
		err := d.Detect(context.Background(), q)
		if err == nil {
			t.Errorf("expected %q to be blocked", q)
		}
	}
}

