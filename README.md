# 🚂 Railguard

[![GitHub tag](https://img.shields.io/github/v/tag/RasmusHilmar1/railguard?include_prereleases)](https://github.com/RasmusHilmar1/railguard/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/RasmusHilmar1/railguard.svg)](https://pkg.go.dev/github.com/RasmusHilmar1/railguard)
[![Go Report Card](https://goreportcard.com/badge/github.com/RasmusHilmar1/railguard)](https://goreportcard.com/report/github.com/RasmusHilmar1/railguard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Railguard** is a Go library that provides safety and validation for LLM (Large Language Model) interactions. It acts as a protective layer between your application and LLM providers, offering prompt injection detection, domain restriction, output validation, and structured response parsing.

## Features

- 🛡️ **Prompt Injection Detection** - Built-in detectors for common attack patterns
- 🎯 **Domain Restriction** - Keep your LLM focused on your use case (e.g., only answer invoice questions)
- ✅ **Output Validation** - Validate JSON, length limits, and custom rules
- 📦 **Schema Parsing** - Parse LLM output into strongly-typed Go structs
- 🔄 **Automatic Retries** - Configurable retry logic with exponential backoff
- 🔌 **Provider Agnostic** - Works with any LLM through a simple interface
- 🎯 **Zero Dependencies** - Only uses the Go standard library

## Installation

```bash
go get github.com/RasmusHilmar1/railguard
```

Requires Go 1.21 or later.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/RasmusHilmar1/railguard"
    "github.com/RasmusHilmar1/railguard/detectors"
    "github.com/RasmusHilmar1/railguard/validators"
)

// Define your response structure
type Response struct {
    Message string `json:"message"`
    Score   int    `json:"score"`
}

func main() {
    // Create a client (implement your own or use ClientFunc)
    client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
        // Call your LLM API here (OpenAI, Anthropic, etc.)
        return `{"message": "Hello!", "score": 95}`, nil
    })

    // Create a guard with safety features
    guard, err := railguard.New(
        railguard.WithClient(client),
        railguard.WithSchema(&Response{}),
        railguard.WithDetectors(
            detectors.NewKeywords(),  // Detect injection keywords
            detectors.NewRole(),      // Detect role manipulation
        ),
        railguard.WithValidators(
            validators.NewJSON(),           // Validate JSON syntax
            validators.NewMaxLength(10000), // Limit output length
        ),
        railguard.WithMaxRetries(3),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Run the pipeline
    result, err := guard.Run(context.Background(), "Tell me a joke")
    if err != nil {
        log.Fatal(err)
    }

    // Access the parsed response
    resp := result.Parsed.(*Response)
    fmt.Printf("Message: %s (Score: %d)\n", resp.Message, resp.Score)
    fmt.Printf("Completed in %d attempts\n", result.Metadata.Attempts)
}
```

## Core Concepts

### Pipeline Architecture

Railguard implements a pipeline with four stages:

```
┌──────────┐    ┌────────────┐    ┌────────────┐    ┌────────┐
│ Detectors │ → │ Generation │ → │ Validators │ → │ Schema │
└──────────┘    └────────────┘    └────────────┘    └────────┘
     ↓               ↓                 ↓
   Fail Fast     Retryable         Retryable
```

1. **Detectors** - Pre-generation safety checks (fail fast, no retry)
2. **Generation** - Call the LLM client (retryable)
3. **Validators** - Post-generation validation (retryable)
4. **Schema** - Parse into Go types (retryable)

### Client Interface

Implement the `Client` interface to connect any LLM:

```go
type Client interface {
    Generate(ctx context.Context, prompt string) (string, error)
}
```

Or use the `ClientFunc` adapter for simple cases:

```go
client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
    // Your LLM API call here
    return response, nil
})
```

---

## Detectors

Detectors examine prompts before they're sent to the LLM. If a detector returns an error, the request fails immediately (no retries).

### Keywords Detector

Detects common prompt injection keywords:

```go
// Use default injection keywords (30+ patterns)
keywords := detectors.NewKeywords()

// Or provide custom keywords
keywords := detectors.NewKeywords("confidential", "secret", "hack")

// Add more keywords to existing detector
keywords := detectors.NewKeywords().WithKeywords("custom", "words")
```

**Default keywords include:** "ignore previous", "disregard instructions", "system prompt", "jailbreak", etc.

### Role Detector

Detects attempts to manipulate the LLM's role or identity:

```go
role := detectors.NewRole()
```

**Detects patterns like:**
- "You are now..."
- "Act as..."
- "Pretend to be..."
- "Enable developer mode"
- "Remove all restrictions"

### Domain Detector (Keyword-based)

Restrict queries to a specific domain using keywords:

```go
// Block off-topic queries with a blocklist
domain := detectors.NewDomain("invoice-search",
    detectors.WithBlockedKeywords("weather", "recipe", "movie", "joke"),
)

// Or require on-topic keywords (stricter)
domain := detectors.NewDomain("invoice-search",
    detectors.WithAllowedKeywords("invoice", "payment", "billing", "receipt"),
    detectors.WithRequireAllowed(true), // Must contain at least one allowed keyword
)

// Combine both approaches
domain := detectors.NewDomain("invoice-search",
    detectors.WithAllowedKeywords("invoice", "payment", "billing"),
    detectors.WithBlockedKeywords(detectors.CommonOffTopicKeywords()...),
    detectors.WithBlockedPatterns(detectors.CommonOffTopicPatterns()...),
    detectors.WithRequireAllowed(true),
)
```

### Intent Detector (LLM-based) ⭐

Use an LLM to intelligently determine if queries are on-topic. This is more flexible than keyword matching:

```go
// Create an intent detector - just describe your domain!
intent := detectors.NewIntent(client, "invoice search",
    detectors.WithDescription("A system for searching invoices, payments, and billing"),
    detectors.WithExamples(
        "Find all unpaid invoices",
        "Show payments from last month",
        "What's the total for customer ABC?",
    ),
)

guard, _ := railguard.New(
    railguard.WithClient(client),
    railguard.WithDetectors(intent),
)

// ✅ Allowed: "Find invoices over $1000"
// ✅ Allowed: "Payment history for ACME Corp"
// ❌ Blocked: "What's the weather?"
// ❌ Blocked: "Tell me a joke"
```

The LLM understands context, so it handles edge cases better than keywords.

### Custom Detector

Create your own detector with `DetectorFunc`:

```go
custom := railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
    if len(prompt) > 10000 {
        return errors.New("prompt too long")
    }
    if containsPII(prompt) {
        return errors.New("prompt contains PII")
    }
    return nil
})
```

---

## Validators

Validators examine LLM output after generation. Validation failures can be retried.

### JSON Validator

Validate that output is valid JSON:

```go
json := validators.NewJSON()                  // Any valid JSON
json := validators.NewJSON().RequireObject()  // Must be JSON object {}
json := validators.NewJSON().RequireArray()   // Must be JSON array []
```

### JSON Extractor

Extract JSON from markdown code blocks (useful when LLMs wrap responses):

```go
extractor := validators.NewJSONExtractor()

// Handles responses like:
// ```json
// {"result": "success"}
// ```
```

### Length Validators

Enforce output length limits:

```go
// By bytes (default)
maxLen := validators.NewMaxLength(1000)           // Max 1000 bytes
minLen := validators.NewMinLength(10)             // Min 10 bytes
rangeLen := validators.NewLengthRange(10, 1000)   // Between 10-1000 bytes

// By characters (runes) - important for Unicode
maxLen := validators.NewMaxLength(1000).WithRunes()   // Max 1000 characters
minLen := validators.NewMinLength(10).WithRunes()
rangeLen := validators.NewLengthRange(10, 1000).WithRunes()
```

### Custom Validator

Create your own validator with `ValidatorFunc`:

```go
custom := railguard.ValidatorFunc(func(ctx context.Context, output string) error {
    if strings.Contains(output, "error") {
        return errors.New("output contains error")
    }
    return nil
})
```

---

## Schema Validation

Schemas validate JSON structure and parse into Go types:

```go
type Response struct {
    Name  string `json:"name"`
    Count int    `json:"count"`
}

guard, _ := railguard.New(
    railguard.WithClient(client),
    railguard.WithSchema(&Response{}),
)

result, _ := guard.Run(ctx, prompt)
resp := result.Parsed.(*Response) // Strongly typed!
```

Schemas are **strict by default** - unknown fields are rejected to prevent hallucinated data.

```go
// Disable strict mode if needed
guard, _ := railguard.New(
    railguard.WithClient(client),
    railguard.WithSchema(&Response{}),
    railguard.WithStrictSchema(false), // Allow unknown fields
)
```

---

## Retry Configuration

Configure retry behavior for transient failures:

```go
config := railguard.RetryConfig{
    MaxAttempts:  5,                      // Total attempts (including first)
    InitialDelay: 100 * time.Millisecond, // Delay before first retry
    MaxDelay:     5 * time.Second,        // Maximum delay cap
    Multiplier:   2.0,                    // Exponential backoff factor
    Jitter:       0.1,                    // ±10% randomization
}

guard, _ := railguard.New(
    railguard.WithClient(client),
    railguard.WithRetry(config),
)

// Or just set max retries with defaults
guard, _ := railguard.New(
    railguard.WithClient(client),
    railguard.WithMaxRetries(5),
)
```

---

## Error Handling

Railguard provides typed errors for different failure modes:

```go
result, err := guard.Run(ctx, prompt)
if err != nil {
    var detErr *railguard.DetectionError
    var valErr *railguard.ValidationError
    var schErr *railguard.SchemaError
    var genErr *railguard.GenerationError
    var maxErr *railguard.MaxRetriesError

    switch {
    case errors.As(err, &detErr):
        log.Printf("Prompt rejected by %s: %v", detErr.Detector, detErr.Err)
    case errors.As(err, &valErr):
        log.Printf("Output rejected by %s: %v", valErr.Validator, valErr.Err)
    case errors.As(err, &schErr):
        log.Printf("Schema validation failed: %v", schErr.Err)
    case errors.As(err, &genErr):
        log.Printf("LLM generation failed: %v", genErr.Err)
    case errors.As(err, &maxErr):
        log.Printf("Max retries (%d) exceeded: %v", maxErr.Attempts, maxErr.LastErr)
    }
}
```

---

## Complete Example: Invoice Search System

Here's a complete example of restricting an LLM to only handle invoice-related queries:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/RasmusHilmar1/railguard"
    "github.com/RasmusHilmar1/railguard/detectors"
    "github.com/RasmusHilmar1/railguard/validators"
)

type InvoiceResponse struct {
    Invoices []Invoice `json:"invoices"`
    Total    float64   `json:"total"`
}

type Invoice struct {
    ID     string  `json:"id"`
    Amount float64 `json:"amount"`
    Status string  `json:"status"`
}

func main() {
    client := yourLLMClient // Your OpenAI/Anthropic/etc client

    // Smart domain restriction using LLM
    invoiceIntent := detectors.NewIntent(client, "invoice search",
        detectors.WithDescription("Search and query invoices, payments, billing, and financial transactions"),
        detectors.WithExamples(
            "Find unpaid invoices",
            "Show payments from last month",
            "Total billing for customer X",
        ),
    )

    guard, err := railguard.New(
        railguard.WithClient(client),
        railguard.WithSchema(&InvoiceResponse{}),
        railguard.WithDetectors(
            detectors.NewKeywords(), // Block prompt injection
            detectors.NewRole(),     // Block role manipulation
            invoiceIntent,           // Only allow invoice queries
        ),
        railguard.WithValidators(
            validators.NewJSON().RequireObject(),
            validators.NewMaxLength(50000),
        ),
        railguard.WithMaxRetries(3),
        railguard.WithTimeout(30 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    // This works ✅
    result, err := guard.Run(context.Background(), "Find all unpaid invoices over $1000")
    if err != nil {
        log.Fatal(err)
    }
    resp := result.Parsed.(*InvoiceResponse)
    fmt.Printf("Found %d invoices, total: $%.2f\n", len(resp.Invoices), resp.Total)

    // This is blocked ❌
    _, err = guard.Run(context.Background(), "What's the weather today?")
    // Error: off-topic: query is not related to invoice search
}
```

---

## API Reference

### Options

| Option | Description |
|--------|-------------|
| `WithClient(Client)` | Set the LLM client (required) |
| `WithSchema(interface{})` | Set the response schema for parsing |
| `WithDetectors(...Detector)` | Add pre-generation detectors |
| `WithValidators(...Validator)` | Add post-generation validators |
| `WithRetry(RetryConfig)` | Set custom retry configuration |
| `WithMaxRetries(int)` | Set max retry attempts |
| `WithTimeout(time.Duration)` | Set operation timeout |
| `WithStrictSchema(bool)` | Enable/disable strict schema mode |

### Built-in Detectors

| Detector | Description |
|----------|-------------|
| `NewKeywords()` | Detect prompt injection keywords |
| `NewRole()` | Detect role manipulation attempts |
| `NewDomain(name, opts...)` | Keyword-based domain restriction |
| `NewIntent(client, domain, opts...)` | LLM-based smart domain restriction |

### Built-in Validators

| Validator | Description |
|-----------|-------------|
| `NewJSON()` | Validate JSON syntax |
| `NewJSONExtractor()` | Extract JSON from markdown blocks |
| `NewMaxLength(n)` | Enforce maximum length |
| `NewMinLength(n)` | Enforce minimum length |
| `NewLengthRange(min, max)` | Enforce length range |

### Result

```go
type Result struct {
    Raw      string      // Raw LLM output
    Parsed   interface{} // Parsed struct (if schema configured)
    Metadata Metadata    // Execution metadata
}

type Metadata struct {
    Attempts int           // Number of attempts made
    Duration time.Duration // Total execution time
}
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
