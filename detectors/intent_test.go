package detectors_test

import (
	"context"
	"testing"

	"github.com/RasmusHilmar1/railguard"
	"github.com/RasmusHilmar1/railguard/detectors"
)

// mockClassifierClient simulates an LLM that classifies queries
type mockClassifierClient struct {
	onTopicQueries map[string]bool
}

func (m *mockClassifierClient) Generate(ctx context.Context, prompt string) (string, error) {
	// Simple mock: check if any on-topic keyword is in the classification prompt
	for keyword, isOnTopic := range m.onTopicQueries {
		if contains(prompt, keyword) {
			if isOnTopic {
				return `{"on_topic": true, "reason": "query is related to invoices"}`, nil
			}
			return `{"on_topic": false, "reason": "query is about weather, not invoices"}`, nil
		}
	}
	return `{"on_topic": true, "reason": "unknown query, allowing"}`, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestIntent(t *testing.T) {
	client := &mockClassifierClient{
		onTopicQueries: map[string]bool{
			"invoice":   true,
			"payment":   true,
			"weather":   false,
			"joke":      false,
			"recipe":    false,
		},
	}

	intent := detectors.NewIntent(client, "invoice search",
		detectors.WithDescription("A system for searching invoices and payments"),
		detectors.WithExamples(
			"Find invoices from last month",
			"Show unpaid bills",
		),
	)

	t.Run("allows on-topic queries", func(t *testing.T) {
		err := intent.Detect(context.Background(), "Find invoice #12345")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("blocks off-topic queries", func(t *testing.T) {
		err := intent.Detect(context.Background(), "What's the weather?")
		if err == nil {
			t.Error("expected error for off-topic query")
		}
	})

	t.Run("name includes domain", func(t *testing.T) {
		if intent.Name() != "intent:invoice search" {
			t.Errorf("unexpected name: %s", intent.Name())
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := intent.Detect(ctx, "test")
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestIntentWithRealPrompt(t *testing.T) {
	// Test that the classification prompt is built correctly
	mockClient := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
		// Verify the prompt contains expected elements
		if !contains(prompt, "invoice search") {
			t.Error("prompt should contain domain name")
		}
		if !contains(prompt, "billing system") {
			t.Error("prompt should contain description")
		}
		if !contains(prompt, "Find invoices") {
			t.Error("prompt should contain examples")
		}
		return `{"on_topic": true, "reason": "test"}`, nil
	})

	intent := detectors.NewIntent(mockClient, "invoice search",
		detectors.WithDescription("A billing system for managing invoices"),
		detectors.WithExamples("Find invoices from Q1"),
	)

	_ = intent.Detect(context.Background(), "test query")
}

