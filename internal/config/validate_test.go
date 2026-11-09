package config_test

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/thomas-vilte/baqueano/internal/config"
)

func TestValidateValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(*config.Config)
	}{
		{name: "defaults"},
		{
			name: "embedded_prompt",
			setup: func(c *config.Config) {
				c.Prompt.SystemFile = ""
			},
		},
		{
			name: "disabled_stackexchange_without_sites",
			setup: func(c *config.Config) {
				c.Sources.StackExchange.Enabled = false
				c.Sources.StackExchange.Sites = nil
			},
		},
		{
			name: "distinct_external_sources",
			setup: func(c *config.Config) {
				c.Sources.External = []config.External{
					{Name: "first", Command: "first-mcp"},
					{Name: "second", Command: "second-mcp"},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			if tt.setup != nil {
				tt.setup(&c)
			}
			secrets := config.Secrets{AnthropicAPIKey: "synthetic-anthropic"}

			if err := c.Validate(secrets, nil); err != nil {
				t.Fatalf("Validate() returned an unexpected error: %v", err)
			}
		})
	}
}

func TestValidateInvalidFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(*config.Config)
		field string
	}{
		{
			name:  "empty_research_model",
			setup: func(c *config.Config) { c.Model.Research = "" },
			field: "model.research",
		},
		{
			name:  "blank_report_model",
			setup: func(c *config.Config) { c.Model.Report = " \t\n" },
			field: "model.report",
		},
		{
			name:  "blank_search_window",
			setup: func(c *config.Config) { c.Search.Since = " " },
			field: "search.since",
		},
		{
			name:  "record_disabled_brave",
			setup: func(c *config.Config) { c.Sources.Brave.Record = true },
			field: "sources.brave.record",
		},
		{
			name: "record_enabled_brave",
			setup: func(c *config.Config) {
				c.Sources.Brave.Enabled = true
				c.Sources.Brave.Record = true
			},
			field: "sources.brave.record",
		},
		{
			name:  "empty_stackexchange_sites",
			setup: func(c *config.Config) { c.Sources.StackExchange.Sites = nil },
			field: "sources.stackexchange.sites",
		},
		{
			name: "blank_stackexchange_site",
			setup: func(c *config.Config) {
				c.Sources.StackExchange.Sites = []string{"stackoverflow", " "}
			},
			field: "sources.stackexchange.sites[1]",
		},
		{
			name: "empty_discourse_forums",
			setup: func(c *config.Config) {
				c.Sources.Discourse.Enabled = true
				c.Sources.Discourse.Forums = nil
			},
			field: "sources.discourse.forums",
		},
		{
			name: "blank_external_name",
			setup: func(c *config.Config) {
				c.Sources.External = []config.External{
					{Name: " ", Command: "forum-mcp"},
				}
			},
			field: "sources.external[0].name",
		},
		{
			name: "blank_external_command",
			setup: func(c *config.Config) {
				c.Sources.External = []config.External{
					{Name: "forum", Command: " "},
				}
			},
			field: "sources.external[0].command",
		},
		{
			name: "duplicate_external_names",
			setup: func(c *config.Config) {
				c.Sources.External = []config.External{
					{Name: "forum", Command: "first-mcp"},
					{Name: "forum", Command: "second-mcp"},
				}
			},
			field: "sources.external[1].name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			tt.setup(&c)
			secrets := config.Secrets{
				AnthropicAPIKey: "synthetic-anthropic",
				BraveAPIKey:     "synthetic-brave",
			}

			requireValidationFields(t, c.Validate(secrets, nil), tt.field)
		})
	}
}

func TestValidatePositiveLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field string
		set   func(*config.Config, int)
	}{
		{"model.max_output_tokens", func(c *config.Config, n int) { c.Model.MaxOutputTokens = n }},
		{"budgets.max_input_tokens", func(c *config.Config, n int) { c.Budgets.MaxInputTokens = n }},
		{"budgets.timeout", func(c *config.Config, n int) { c.Budgets.Timeout = time.Duration(n) }},
		{"budgets.tool_timeout", func(c *config.Config, n int) { c.Budgets.ToolTimeout = time.Duration(n) }},
		{"budgets.max_results_per_tool", func(c *config.Config, n int) { c.Budgets.MaxResultsPerTool = n }},
		{"cache.ttl", func(c *config.Config, n int) { c.Cache.TTL = time.Duration(n) }},
	}

	for _, tt := range tests {
		for _, value := range []int{-1, 0, 1} {
			t.Run(fmt.Sprintf("%s/%d", tt.field, value), func(t *testing.T) {
				t.Parallel()

				c := config.Defaults()
				tt.set(&c, value)
				secrets := config.Secrets{AnthropicAPIKey: "synthetic-anthropic"}

				err := c.Validate(secrets, nil)
				if value > 0 {
					if err != nil {
						t.Fatalf("Validate() returned an unexpected error: %v", err)
					}
					return
				}
				requireValidationFields(t, err, tt.field)
			})
		}
	}
}

func TestValidateMaxSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		steps int
		valid bool
	}{
		{-1, false},
		{0, false},
		{1, true},
		{8, true},
		{20, true},
		{21, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.steps), func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			c.Budgets.MaxSteps = tt.steps
			secrets := config.Secrets{AnthropicAPIKey: "synthetic-anthropic"}

			err := c.Validate(secrets, nil)
			if tt.valid {
				if err != nil {
					t.Fatalf("Validate() returned an unexpected error: %v", err)
				}
				return
			}
			requireValidationFields(t, err, "budgets.max_steps")
		})
	}
}

func TestValidateMaxCostUSD(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cost  float64
		valid bool
	}{
		{"zero", 0, false},
		{"negative", -0.10, false},
		{"nan", math.NaN(), false},
		{"positive_infinity", math.Inf(1), false},
		{"negative_infinity", math.Inf(-1), false},
		{"positive", 0.10, true},
		{"small_positive", math.SmallestNonzeroFloat64, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			c.Budgets.MaxCostUSD = tt.cost
			secrets := config.Secrets{AnthropicAPIKey: "synthetic-anthropic"}

			err := c.Validate(secrets, nil)
			if tt.valid {
				if err != nil {
					t.Fatalf("Validate() returned an unexpected error: %v", err)
				}
				return
			}
			requireValidationFields(t, err, "budgets.max_cost_usd")
		})
	}
}

func TestValidateCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		github  bool
		brave   bool
		secrets config.Secrets
		field   string
	}{
		{
			name:  "missing_anthropic",
			field: "ANTHROPIC_API_KEY",
		},
		{
			name:    "blank_anthropic",
			secrets: config.Secrets{AnthropicAPIKey: " \t"},
			field:   "ANTHROPIC_API_KEY",
		},
		{
			name:    "missing_github",
			github:  true,
			secrets: config.Secrets{AnthropicAPIKey: "synthetic-anthropic"},
			field:   "GITHUB_TOKEN",
		},
		{
			name:    "blank_github",
			github:  true,
			secrets: config.Secrets{AnthropicAPIKey: "synthetic-anthropic", GitHubToken: " "},
			field:   "GITHUB_TOKEN",
		},
		{
			name:    "missing_brave",
			brave:   true,
			secrets: config.Secrets{AnthropicAPIKey: "synthetic-anthropic"},
			field:   "BRAVE_API_KEY",
		},
		{
			name:    "blank_brave",
			brave:   true,
			secrets: config.Secrets{AnthropicAPIKey: "synthetic-anthropic", BraveAPIKey: " "},
			field:   "BRAVE_API_KEY",
		},
		{
			name:    "optional_keys_absent",
			secrets: config.Secrets{AnthropicAPIKey: "synthetic-anthropic"},
		},
		{
			name:   "enabled_sources_with_keys",
			github: true,
			brave:  true,
			secrets: config.Secrets{
				AnthropicAPIKey: "synthetic-anthropic",
				GitHubToken:     "synthetic-github",
				BraveAPIKey:     "synthetic-brave",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			c.Sources.GitHub.Enabled = tt.github
			c.Sources.Brave.Enabled = tt.brave

			err := c.Validate(tt.secrets, nil)
			if tt.field == "" {
				if err != nil {
					t.Fatalf("Validate() returned an unexpected error: %v", err)
				}
				return
			}
			requireValidationFields(t, err, tt.field)
		})
	}
}

func TestValidateForumAllowlist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		forum   string
		allowed []string
		valid   bool
	}{
		{"approved", "forum.example", []string{"forum.example"}, true},
		{"uppercase_forum", "FORUM.EXAMPLE", []string{"forum.example"}, true},
		{"uppercase_allowlist", "forum.example", []string{"FORUM.EXAMPLE"}, true},
		{"missing_allowlist", "forum.example", nil, false},
		{"unknown", "other.example", []string{"forum.example"}, false},
		{"subdomain", "sub.forum.example", []string{"forum.example"}, false},
		{"misleading_suffix", "forum.example.evil.test", []string{"forum.example"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			c.Sources.Discourse.Enabled = true
			c.Sources.Discourse.Forums = []string{tt.forum}
			secrets := config.Secrets{AnthropicAPIKey: "synthetic-anthropic"}

			err := c.Validate(secrets, tt.allowed)
			if tt.valid {
				if err != nil {
					t.Fatalf("Validate() returned an unexpected error: %v", err)
				}
				return
			}
			requireValidationFields(t, err, "sources.discourse.forums[0]")
		})
	}
}

func TestValidateRejectsMalformedForums(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		forum string
	}{
		{"empty", ""},
		{"url", "https://forum.example"},
		{"port", "forum.example:443"},
		{"path", "forum.example/topics"},
		{"query", "forum.example?search=go"},
		{"fragment", "forum.example#topic"},
		{"userinfo", "user@forum.example"},
		{"leading_space", " forum.example"},
		{"trailing_space", "forum.example "},
		{"unicode_space", "forum.\u00a0example"},
		{"invalid_escape", "forum%zz.example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Defaults()
			c.Sources.Discourse.Enabled = true
			c.Sources.Discourse.Forums = []string{tt.forum}
			secrets := config.Secrets{AnthropicAPIKey: "synthetic-anthropic"}

			// Approval does not make a URL, port or path a bare hostname.
			allowed := []string{tt.forum}
			requireValidationFields(t, c.Validate(secrets, allowed),
				"sources.discourse.forums[0]")
		})
	}
}

func TestValidateAccumulatesErrors(t *testing.T) {
	t.Parallel()

	c := config.Defaults()
	c.Model.Research = ""
	c.Budgets.Timeout = 0
	c.Sources.GitHub.Enabled = true
	c.Sources.Brave.Record = true
	c.Sources.StackExchange.Sites = []string{"", " "}

	requireValidationFields(t, c.Validate(config.Secrets{}, nil),
		"model.research",
		"budgets.timeout",
		"ANTHROPIC_API_KEY",
		"GITHUB_TOKEN",
		"sources.brave.record",
		"sources.stackexchange.sites[0]",
		"sources.stackexchange.sites[1]",
	)
}

func TestValidateDoesNotExposeValues(t *testing.T) {
	t.Parallel()

	c := config.Defaults()
	c.Sources.Discourse.Enabled = true
	c.Sources.Discourse.Forums = []string{"private-forum.example"}
	c.Sources.External = []config.External{
		{Name: "private-plugin", Command: "first-mcp"},
		{Name: "private-plugin", Command: "second-mcp"},
	}
	secrets := config.Secrets{
		AnthropicAPIKey:  "synthetic-anthropic",
		StackExchangeKey: "synthetic-stackexchange",
		GitHubToken:      "synthetic-github",
		BraveAPIKey:      "synthetic-brave",
	}

	err := c.Validate(secrets, nil)
	requireValidationFields(t, err,
		"sources.discourse.forums[0]",
		"sources.external[1].name",
	)

	for _, value := range []string{
		secrets.AnthropicAPIKey,
		secrets.StackExchangeKey,
		secrets.GitHubToken,
		secrets.BraveAPIKey,
		c.Sources.Discourse.Forums[0],
		c.Sources.External[0].Name,
	} {
		if strings.Contains(err.Error(), value) {
			t.Error("Validate() exposed a configured value")
		}
	}
}

func TestValidateNilConfig(t *testing.T) {
	t.Parallel()

	var c *config.Config

	requireValidationFields(t, c.Validate(config.Secrets{}, nil), "config")
}

func requireValidationFields(t *testing.T, err error, want ...string) {
	t.Helper()

	if err == nil {
		t.Fatalf("Validate() returned nil; want fields %v", want)
	}

	problems := []error{err}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		problems = joined.Unwrap()
	}

	got := make([]string, 0, len(problems))
	for _, problem := range problems {
		var fieldErr *config.FieldError
		if !errors.As(problem, &fieldErr) {
			t.Fatalf("Validate() returned %T; want *config.FieldError", problem)
		}
		got = append(got, fieldErr.Field)
	}

	slices.Sort(got)
	want = slices.Clone(want)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Errorf("Validate() fields = %v; want %v", got, want)
	}
}
