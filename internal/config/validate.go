package config

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"slices"
	"strings"
	"unicode"
)

// FieldError identifies an invalid field without exposing its value.
type FieldError struct {
	Field  string
	Reason string
}

// Error returns a message safe to include in startup diagnostics.
func (e *FieldError) Error() string {
	return "config: " + e.Field + ": " + e.Reason
}

// Validate checks settings, credentials and approved forum hosts.
func (c *Config) Validate(secrets Secrets, allowedHosts []string) error {
	if c == nil {
		return &FieldError{
			Field:  "config",
			Reason: "must not be nil",
		}
	}

	var problems []error

	check := func(valid bool, field, reason string) {
		if !valid {
			problems = append(problems, &FieldError{
				Field:  field,
				Reason: reason,
			})
		}
	}

	check(notBlank(c.Model.Research),
		"model.research", "must not be empty")
	check(notBlank(c.Model.Report),
		"model.report", "must not be empty")
	check(c.Model.MaxOutputTokens > 0,
		"model.max_output_tokens", "must be greater than zero")
	check(notBlank(c.Search.Since),
		"search.since", "must not be empty")

	check(c.Budgets.MaxSteps >= 1 && c.Budgets.MaxSteps <= 20,
		"budgets.max_steps", "must be between 1 and 20")
	check(c.Budgets.MaxInputTokens > 0,
		"budgets.max_input_tokens", "must be greater than zero")
	check(c.Budgets.Timeout > 0,
		"budgets.timeout", "must be greater than zero")
	check(c.Budgets.ToolTimeout > 0,
		"budgets.tool_timeout", "must be greater than zero")
	check(c.Budgets.MaxResultsPerTool > 0,
		"budgets.max_results_per_tool", "must be greater than zero")

	cost := c.Budgets.MaxCostUSD
	check(cost > 0 && !math.IsNaN(cost) && !math.IsInf(cost, 0),
		"budgets.max_cost_usd", "must be finite and greater than zero")

	check(c.Cache.TTL > 0,
		"cache.ttl", "must be greater than zero")

	check(notBlank(secrets.AnthropicAPIKey),
		"ANTHROPIC_API_KEY", "is required")

	if c.Sources.GitHub.Enabled {
		check(notBlank(secrets.GitHubToken),
			"GITHUB_TOKEN", "is required when GitHub is enabled")
	}

	if c.Sources.Brave.Enabled {
		check(notBlank(secrets.BraveAPIKey),
			"BRAVE_API_KEY", "is required when Brave is enabled")
	}

	check(!c.Sources.Brave.Record,
		"sources.brave.record", "must be false")

	if c.Sources.StackExchange.Enabled {
		check(len(c.Sources.StackExchange.Sites) > 0,
			"sources.stackexchange.sites", "must contain at least one site")

		for i, site := range c.Sources.StackExchange.Sites {
			check(notBlank(site),
				fmt.Sprintf("sources.stackexchange.sites[%d]", i),
				"must not be empty")
		}
	}

	if c.Sources.Discourse.Enabled {
		check(len(c.Sources.Discourse.Forums) > 0,
			"sources.discourse.forums", "must contain at least one forum")

		for i, forum := range c.Sources.Discourse.Forums {
			check(allowedForum(forum, allowedHosts),
				fmt.Sprintf("sources.discourse.forums[%d]", i),
				"must be a hostname in the approved list")
		}
	}

	names := make(map[string]struct{}, len(c.Sources.External))

	for i, source := range c.Sources.External {
		field := fmt.Sprintf("sources.external[%d]", i)

		check(notBlank(source.Name),
			field+".name", "must not be empty")
		check(notBlank(source.Command),
			field+".command", "must not be empty")

		if notBlank(source.Name) {
			_, exists := names[source.Name]
			check(!exists, field+".name", "must be unique")
			names[source.Name] = struct{}{}
		}
	}

	return errors.Join(problems...)
}

func notBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func allowedForum(forum string, allowedHosts []string) bool {
	if forum == "" || strings.IndexFunc(forum, unicode.IsSpace) >= 0 {
		return false
	}

	parsed, err := url.Parse("https://" + forum)
	if err != nil || parsed.Hostname() != forum {
		return false
	}

	return slices.ContainsFunc(allowedHosts, func(host string) bool {
		return strings.EqualFold(forum, host)
	})
}
