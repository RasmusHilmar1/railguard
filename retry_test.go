package railguard_test

import (
	"testing"
	"time"

	"github.com/RasmusHilmar1/railguard"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := railguard.DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", config.MaxAttempts)
	}
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected InitialDelay 100ms, got %v", config.InitialDelay)
	}
	if config.MaxDelay != 5*time.Second {
		t.Errorf("expected MaxDelay 5s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %v", config.Multiplier)
	}
	if config.Jitter != 0.1 {
		t.Errorf("expected Jitter 0.1, got %v", config.Jitter)
	}
}

func TestRetryConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  railguard.RetryConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  railguard.DefaultRetryConfig(),
			wantErr: false,
		},
		{
			name: "zero max attempts",
			config: railguard.RetryConfig{
				MaxAttempts:  0,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       0.1,
			},
			wantErr: true,
		},
		{
			name: "negative max attempts",
			config: railguard.RetryConfig{
				MaxAttempts:  -1,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       0.1,
			},
			wantErr: true,
		},
		{
			name: "negative initial delay",
			config: railguard.RetryConfig{
				MaxAttempts:  3,
				InitialDelay: -100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       0.1,
			},
			wantErr: true,
		},
		{
			name: "negative max delay",
			config: railguard.RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     -5 * time.Second,
				Multiplier:   2.0,
				Jitter:       0.1,
			},
			wantErr: true,
		},
		{
			name: "multiplier less than 1",
			config: railguard.RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   0.5,
				Jitter:       0.1,
			},
			wantErr: true,
		},
		{
			name: "jitter negative",
			config: railguard.RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       -0.1,
			},
			wantErr: true,
		},
		{
			name: "jitter greater than 1",
			config: railguard.RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       1.5,
			},
			wantErr: true,
		},
		{
			name: "single attempt valid",
			config: railguard.RetryConfig{
				MaxAttempts:  1,
				InitialDelay: 0,
				MaxDelay:     0,
				Multiplier:   1.0,
				Jitter:       0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

