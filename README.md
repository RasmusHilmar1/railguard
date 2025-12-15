# 🚂 Railguard

[![Go Reference](https://pkg.go.dev/badge/github.com/RasmusHilmar1/railguard.svg)](https://pkg.go.dev/github.com/RasmusHilmar1/railguard)
[![Go Report Card](https://goreportcard.com/badge/github.com/RasmusHilmar1/railguard)](https://goreportcard.com/report/github.com/RasmusHilmar1/railguard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Railguard** is a Go library that provides safety and validation for LLM (Large Language Model) interactions. It acts as a protective layer between your application and LLM providers, offering prompt injection detection, output validation, and structured response parsing.

## Features

- 🛡️ **Prompt Injection Detection** - Built-in detectors for common attack patterns
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
            detectors.NewRole(),       // Detect role manipulation
        ),
        railguard.WithValidators(
            validators.NewJSON(),              // Validate JSON syntax
            validators.NewMaxLength(10000),    // Limit output length
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

### Detectors

Detectors examine prompts before they're sent to the LLM:

```go
// Built-in detectors
keywords := detectors.NewKeywords()                    // Default injection keywords
keywords := detectors.NewKeywords("custom", "words")   // Custom keywords
role := detectors.NewRole()                            // Role manipulation patterns

// Custom detector using DetectorFunc
custom := railguard.DetectorFunc(func(ctx context.Context, prompt string) error {
    if containsBadContent(prompt) {
        return errors.New("rejected")
    }
    return nil
})
```

### Validators

Validators examine LLM output before it's returned:

```go
// Built-in validators
json := validators.NewJSON()                           // Any valid JSON
json := validators.NewJSON().RequireObject()           // Must be JSON object
json := validators.NewJSON().RequireArray()            // Must be JSON array
extractor := validators.NewJSONExtractor()             // Extract from markdown blocks

maxLen := validators.NewMaxLength(1000)                // Max 1000 bytes
maxLen := validators.NewMaxLength(1000).WithRunes()    // Max 1000 characters
minLen := validators.NewMinLength(10)                  // Min 10 bytes
range := validators.NewLengthRange(10, 1000)           // Between 10-1000 bytes

// Custom validator using ValidatorFunc
custom := railguard.ValidatorFunc(func(ctx context.Context, output string) error {
    if !isValid(output) {
        return errors.New("invalid output")
    }
    return nil
})
```

### Schema Validation

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

### Retry Configuration

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

### Error Handling

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
