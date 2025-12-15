package detectors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RasmusHilmar1/railguard"
)

// Intent uses an LLM to classify whether a prompt is on-topic for a given domain.
// This is more flexible than keyword matching as it understands semantic meaning.
type Intent struct {
	client      railguard.Client
	domain      string
	description string
	examples    []string
}

// IntentOption configures an Intent detector.
type IntentOption func(*Intent)

// NewIntent creates an Intent detector that uses an LLM to check if queries
// are relevant to the specified domain.
//
// Example:
//
//	intent := detectors.NewIntent(client, "invoice search",
//	    detectors.WithDescription("A system for searching and querying invoices, payments, and billing information"),
//	    detectors.WithExamples(
//	        "Find invoices from last month",
//	        "Show unpaid bills",
//	        "Payment history for customer X",
//	    ),
//	)
func NewIntent(client railguard.Client, domain string, opts ...IntentOption) *Intent {
	i := &Intent{
		client: client,
		domain: domain,
	}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

// WithDescription provides a description of what the domain/system does.
// This helps the LLM understand context for classification.
func WithDescription(desc string) IntentOption {
	return func(i *Intent) {
		i.description = desc
	}
}

// WithExamples provides example on-topic queries.
// These help the LLM understand what valid queries look like.
func WithExamples(examples ...string) IntentOption {
	return func(i *Intent) {
		i.examples = append(i.examples, examples...)
	}
}

// classificationResponse is the expected JSON response from the classifier.
type classificationResponse struct {
	OnTopic bool   `json:"on_topic"`
	Reason  string `json:"reason"`
}

// Detect uses the LLM to check if the prompt is relevant to the domain.
func (i *Intent) Detect(ctx context.Context, prompt string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	classifyPrompt := i.buildClassificationPrompt(prompt)

	response, err := i.client.Generate(ctx, classifyPrompt)
	if err != nil {
		// If classification fails, allow the query through (fail open)
		// You could change this to fail closed if preferred
		return nil
	}

	// Parse the classification response
	var result classificationResponse
	if err := json.Unmarshal([]byte(extractJSON(response)), &result); err != nil {
		// If parsing fails, allow through
		return nil
	}

	if !result.OnTopic {
		reason := result.Reason
		if reason == "" {
			reason = "query is not related to " + i.domain
		}
		return fmt.Errorf("off-topic: %s", reason)
	}

	return nil
}

// buildClassificationPrompt creates the prompt for the classifier LLM.
func (i *Intent) buildClassificationPrompt(userPrompt string) string {
	var sb strings.Builder

	sb.WriteString("You are a query classifier. Determine if the user's query is relevant to the specified domain.\n\n")

	sb.WriteString("DOMAIN: ")
	sb.WriteString(i.domain)
	sb.WriteString("\n")

	if i.description != "" {
		sb.WriteString("DESCRIPTION: ")
		sb.WriteString(i.description)
		sb.WriteString("\n")
	}

	if len(i.examples) > 0 {
		sb.WriteString("\nEXAMPLES OF VALID QUERIES:\n")
		for _, ex := range i.examples {
			sb.WriteString("- ")
			sb.WriteString(ex)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nUSER QUERY: ")
	sb.WriteString(userPrompt)
	sb.WriteString("\n\n")

	sb.WriteString(`Respond with JSON only:
{"on_topic": true/false, "reason": "brief explanation"}

Rules:
- Return on_topic=true if the query is related to the domain
- Return on_topic=false if the query is about something unrelated (weather, jokes, general knowledge, etc.)
- Be lenient for ambiguous queries that could be domain-related
- The reason should be brief (under 20 words)`)

	return sb.String()
}

// Name returns the detector's name.
func (i *Intent) Name() string {
	return "intent:" + i.domain
}

// extractJSON attempts to extract JSON from a response that may contain markdown.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	// Try to find JSON in markdown code block
	if idx := strings.Index(s, "```json"); idx != -1 {
		start := idx + 7
		if end := strings.Index(s[start:], "```"); end != -1 {
			return strings.TrimSpace(s[start : start+end])
		}
	}

	// Try to find JSON in generic code block
	if idx := strings.Index(s, "```"); idx != -1 {
		start := idx + 3
		// Skip to next line
		if nl := strings.Index(s[start:], "\n"); nl != -1 {
			start += nl + 1
		}
		if end := strings.Index(s[start:], "```"); end != -1 {
			return strings.TrimSpace(s[start : start+end])
		}
	}

	// Try to find raw JSON object
	if start := strings.Index(s, "{"); start != -1 {
		if end := strings.LastIndex(s, "}"); end != -1 && end > start {
			return s[start : end+1]
		}
	}

	return s
}

