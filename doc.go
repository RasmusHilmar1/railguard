/*
Package railguard provides safety and validation for LLM (Large Language Model) interactions.

Railguard acts as a protective layer between your application and LLM providers,
offering prompt injection detection, output validation, and structured response parsing.

# Overview

Railguard implements a pipeline-based approach to LLM safety:

 1. Detection - Pre-generation safety checks (e.g., prompt injection detection)
 2. Generation - Calling the LLM client
 3. Validation - Post-generation output validation (e.g., JSON syntax)
 4. Schema - Structured output parsing into Go types

Detection failures are fatal and not retried. Generation and validation failures
may be retried based on the retry configuration.

# Quick Start

	type Response struct {
	    Message string `json:"message"`
	}

	// Create a guard with your LLM client
	guard, err := railguard.New(
	    railguard.WithClient(myLLMClient),
	    railguard.WithSchema(&Response{}),
	    railguard.WithDetectors(detectors.NewKeywords()),
	    railguard.WithValidators(validators.NewJSON()),
	    railguard.WithMaxRetries(3),
	)

	// Run the pipeline
	result, err := guard.Run(ctx, "Hello, how are you?")
	if err != nil {
	    log.Fatal(err)
	}

	// Access the parsed response
	resp := result.Parsed.(*Response)
	fmt.Println(resp.Message)

# Client Interface

Implement the Client interface to connect any LLM provider:

	type Client interface {
	    Generate(ctx context.Context, prompt string) (string, error)
	}

For simple use cases, use ClientFunc:

	client := railguard.ClientFunc(func(ctx context.Context, prompt string) (string, error) {
	    // Call your LLM API here
	    return response, nil
	})

# Built-in Detectors

The detectors package provides pre-built detectors:

  - Keywords - Detects prompt injection keywords
  - Role - Detects role manipulation attempts

# Built-in Validators

The validators package provides pre-built validators:

  - JSON - Validates JSON syntax
  - JSONExtractor - Extracts JSON from markdown code blocks
  - MaxLength - Enforces output length limits
  - MinLength - Enforces minimum output length
  - LengthRange - Enforces output length within a range

# Schema Validation

Schemas validate that LLM output matches your Go types:

	type Response struct {
	    Name  string `json:"name"`
	    Count int    `json:"count"`
	}

	guard, _ := railguard.New(
	    railguard.WithClient(client),
	    railguard.WithSchema(&Response{}),
	)

By default, schemas are strict and reject unknown fields to prevent
hallucinated data from entering your application.

# Retry Configuration

Control retry behavior with RetryConfig:

	config := railguard.RetryConfig{
	    MaxAttempts:  5,
	    InitialDelay: 100 * time.Millisecond,
	    MaxDelay:     5 * time.Second,
	    Multiplier:   2.0,
	    Jitter:       0.1,
	}

	guard, _ := railguard.New(
	    railguard.WithClient(client),
	    railguard.WithRetry(config),
	)

# Error Types

Railguard provides typed errors for handling different failure modes:

  - DetectionError - A detector rejected the prompt
  - ValidationError - A validator rejected the output
  - SchemaError - The output didn't match the schema
  - GenerationError - The LLM client failed
  - MaxRetriesError - Maximum retries exceeded

Use errors.As to handle specific error types:

	result, err := guard.Run(ctx, prompt)
	if err != nil {
	    var detErr *railguard.DetectionError
	    if errors.As(err, &detErr) {
	        log.Printf("Prompt rejected by %s: %v", detErr.Detector, detErr.Err)
	    }
	}
*/
package railguard

