package railguard

import "context"

// Client represents any LLM that can generate text from a prompt.
// Implementations should handle their own authentication, rate limiting,
// and any provider-specific configuration.
type Client interface {
	// Generate produces a response for the given prompt.
	// The context should be used for cancellation and timeouts.
	// Returns the generated text or an error if generation fails.
	Generate(ctx context.Context, prompt string) (string, error)
}

// ClientFunc is an adapter that allows ordinary functions to be used as Clients.
// This is useful for testing or simple use cases where a full Client implementation
// is not needed.
type ClientFunc func(ctx context.Context, prompt string) (string, error)

// Generate implements the Client interface by calling the function itself.
func (f ClientFunc) Generate(ctx context.Context, prompt string) (string, error) {
	return f(ctx, prompt)
}

