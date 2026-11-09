package config

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
)

func applyEnv(ctx context.Context, c *Config, getenv func(string) string) error {
	if getenv == nil {
		return nil
	}

	bindings := []struct {
		name     string
		expected string
		set      func(string) error
	}{
		{"MODEL_RESEARCH", "a string", envValue(&c.Model.Research, envString)},
		{"MODEL_REPORT", "a string", envValue(&c.Model.Report, envString)},
		{"MODEL_MAX_OUTPUT_TOKENS", "an integer", envValue(&c.Model.MaxOutputTokens, strconv.Atoi)},
		{"SEARCH_SINCE", "a string", envValue(&c.Search.Since, envString)},
		{"BUDGETS_MAX_STEPS", "an integer", envValue(&c.Budgets.MaxSteps, strconv.Atoi)},
		{"BUDGETS_MAX_INPUT_TOKENS", "an integer", envValue(&c.Budgets.MaxInputTokens, strconv.Atoi)},
		{"BUDGETS_TIMEOUT", "a duration", envValue(&c.Budgets.Timeout, time.ParseDuration)},
		{"BUDGETS_TOOL_TIMEOUT", "a duration", envValue(&c.Budgets.ToolTimeout, time.ParseDuration)},
		{"BUDGETS_MAX_RESULTS_PER_TOOL", "an integer", envValue(&c.Budgets.MaxResultsPerTool, strconv.Atoi)},
		{"BUDGETS_MAX_COST_USD", "a number", envValue(&c.Budgets.MaxCostUSD, envFloat)},
		{"PROMPT_SYSTEM_FILE", "a string", envValue(&c.Prompt.SystemFile, envString)},
		{"SOURCES_HACKERNEWS_ENABLED", "a boolean", envValue(&c.Sources.HackerNews.Enabled, strconv.ParseBool)},
		{"SOURCES_STACKEXCHANGE_ENABLED", "a boolean", envValue(&c.Sources.StackExchange.Enabled, strconv.ParseBool)},
		{"SOURCES_STACKEXCHANGE_SITES", "a YAML sequence", envValue(&c.Sources.StackExchange.Sites, envList[string])},
		{"SOURCES_GITHUB_ENABLED", "a boolean", envValue(&c.Sources.GitHub.Enabled, strconv.ParseBool)},
		{"SOURCES_BRAVE_ENABLED", "a boolean", envValue(&c.Sources.Brave.Enabled, strconv.ParseBool)},
		{"SOURCES_BRAVE_RECORD", "a boolean", envValue(&c.Sources.Brave.Record, strconv.ParseBool)},
		{"SOURCES_DISCOURSE_ENABLED", "a boolean", envValue(&c.Sources.Discourse.Enabled, strconv.ParseBool)},
		{"SOURCES_DISCOURSE_FORUMS", "a YAML sequence", envValue(&c.Sources.Discourse.Forums, envList[string])},
		{"SOURCES_EXTERNAL", "a YAML sequence", envValue(&c.Sources.External, envList[External])},
		{"CACHE_TTL", "a duration", envValue(&c.Cache.TTL, time.ParseDuration)},
		{"MEMORY_SESSION", "a boolean", envValue(&c.Memory.Session, strconv.ParseBool)},
		{"MEMORY_LONG_TERM", "a boolean", envValue(&c.Memory.LongTerm, strconv.ParseBool)},
		{"TELEMETRY_CAPTURE_CONTENT", "a boolean", envValue(&c.Telemetry.CaptureContent, strconv.ParseBool)},
	}

	var problems []error

	for _, binding := range bindings {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("config: read environment: %w", err)
		}

		key := "BAQUEANO_" + binding.name
		value := getenv(key)
		if value == "" {
			continue
		}

		if err := binding.set(value); err != nil {
			// Parser errors include the input, which may contain credentials.
			problems = append(problems, &FieldError{
				Field:  key,
				Reason: "must be " + binding.expected,
			})
		}
	}

	return errors.Join(problems...)
}

func envValue[T any](target *T, parse func(string) (T, error)) func(string) error {
	return func(value string) error {
		parsed, err := parse(value)
		if err != nil {
			return err
		}
		*target = parsed
		return nil
	}
}

func envString(value string) (string, error) {
	return value, nil
}

func envFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func envList[T any](value string) ([]T, error) {
	var items []T
	if err := decodeYAML([]byte(value), &items); err != nil {
		return nil, err
	}
	if items == nil {
		return nil, errors.New("expected a YAML sequence")
	}
	return items, nil
}
