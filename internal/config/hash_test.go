package config_test

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/thomas-vilte/baqueano/internal/config"
)

func TestHashKnownValue(t *testing.T) {
	t.Parallel()

	var c config.Config

	got, err := c.Hash()
	if err != nil {
		t.Fatalf("Hash() returned an unexpected error: %v", err)
	}

	const want = "sha256:f955adee8b85"
	if got != want {
		t.Errorf("Hash() = %q; want %q", got, want)
	}
}

func TestHashDeterministic(t *testing.T) {
	t.Parallel()

	c := config.Defaults()
	first, err := c.Hash()
	if err != nil {
		t.Fatalf("first Hash() failed: %v", err)
	}

	second, err := c.Hash()
	if err != nil {
		t.Fatalf("second Hash() failed: %v", err)
	}

	if first != second {
		t.Errorf("Hash() changed between calls: %q and %q", first, second)
	}
}

func TestHashChangesWithConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*config.Config)
	}{
		{
			name: "research_model",
			change: func(c *config.Config) {
				c.Model.Research = "another-model"
			},
		},
		{
			name: "timeout",
			change: func(c *config.Config) {
				c.Budgets.Timeout += time.Second
			},
		},
		{
			name: "cost_limit",
			change: func(c *config.Config) {
				c.Budgets.MaxCostUSD = 0.20
			},
		},
		{
			name: "enabled_source",
			change: func(c *config.Config) {
				c.Sources.GitHub.Enabled = true
			},
		},
		{
			name: "site_order",
			change: func(c *config.Config) {
				sites := c.Sources.StackExchange.Sites
				sites[0], sites[1] = sites[1], sites[0]
			},
		},
		{
			name: "telemetry",
			change: func(c *config.Config) {
				c.Telemetry.CaptureContent = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			before, err := c.Hash()
			if err != nil {
				t.Fatalf("Hash() before change failed: %v", err)
			}

			tt.change(&c)
			after, err := c.Hash()
			if err != nil {
				t.Fatalf("Hash() after change failed: %v", err)
			}

			if before == after {
				t.Errorf("Hash() stayed %q after changing %s", after, tt.name)
			}
		})
	}
}

func TestHashIgnoresSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secrets config.Secrets
	}{
		{"anthropic", config.Secrets{AnthropicAPIKey: "synthetic-anthropic"}},
		{"stackexchange", config.Secrets{StackExchangeKey: "synthetic-stackexchange"}},
		{"github", config.Secrets{GitHubToken: "synthetic-github"}},
		{"brave", config.Secrets{BraveAPIKey: "synthetic-brave"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			first := config.Loaded{Config: config.Defaults()}
			second := config.Loaded{
				Config:  config.Defaults(),
				Secrets: tt.secrets,
			}

			firstHash, err := first.Config.Hash()
			if err != nil {
				t.Fatalf("Hash() without secrets failed: %v", err)
			}

			secondHash, err := second.Config.Hash()
			if err != nil {
				t.Fatalf("Hash() with secrets failed: %v", err)
			}

			if firstHash != secondHash {
				t.Errorf("Hash() changed when only %s credentials changed", tt.name)
			}
		})
	}
}

func TestHashRejectsNonFiniteValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cost float64
	}{
		{"nan", math.NaN()},
		{"positive_infinity", math.Inf(1)},
		{"negative_infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			c.Budgets.MaxCostUSD = tt.cost

			got, err := c.Hash()
			if _, ok := errors.AsType[*json.UnsupportedValueError](err); !ok {
				t.Fatalf("Hash() error = %v; want *json.UnsupportedValueError", err)
			}
			if got != "" {
				t.Errorf("Hash() = %q on failure; want an empty string", got)
			}
		})
	}
}

func TestHashNilConfig(t *testing.T) {
	t.Parallel()

	var c *config.Config

	got, err := c.Hash()
	var fieldErr *config.FieldError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("Hash() error = %v; want *config.FieldError", err)
	}
	if fieldErr.Field != "config" {
		t.Errorf("Hash() error field = %q; want config", fieldErr.Field)
	}
	if got != "" {
		t.Errorf("Hash() = %q on failure; want an empty string", got)
	}
}
