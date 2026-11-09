package config_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/thomas-vilte/baqueano/internal/config"
)

func TestLoadEnvironmentBindings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key   string
		value string
		file  string
		want  func(*config.Config)
	}{
		{key: "MODEL_RESEARCH", value: "false", want: func(c *config.Config) { c.Model.Research = "false" }},
		{key: "MODEL_REPORT", value: "model with spaces", want: func(c *config.Config) { c.Model.Report = "model with spaces" }},
		{key: "MODEL_MAX_OUTPUT_TOKENS", value: "2500", want: func(c *config.Config) { c.Model.MaxOutputTokens = 2500 }},
		{key: "SEARCH_SINCE", value: "1y", want: func(c *config.Config) { c.Search.Since = "1y" }},
		{key: "BUDGETS_MAX_STEPS", value: "7", want: func(c *config.Config) { c.Budgets.MaxSteps = 7 }},
		{key: "BUDGETS_MAX_INPUT_TOKENS", value: "50000", want: func(c *config.Config) { c.Budgets.MaxInputTokens = 50000 }},
		{key: "BUDGETS_TIMEOUT", value: "90s", want: func(c *config.Config) { c.Budgets.Timeout = 90 * time.Second }},
		{key: "BUDGETS_TOOL_TIMEOUT", value: "5s", want: func(c *config.Config) { c.Budgets.ToolTimeout = 5 * time.Second }},
		{key: "BUDGETS_MAX_RESULTS_PER_TOOL", value: "20", want: func(c *config.Config) { c.Budgets.MaxResultsPerTool = 20 }},
		{key: "BUDGETS_MAX_COST_USD", value: "0.25", want: func(c *config.Config) { c.Budgets.MaxCostUSD = 0.25 }},
		{key: "PROMPT_SYSTEM_FILE", value: "C:\\local\\system.md", want: func(c *config.Config) { c.Prompt.SystemFile = "C:\\local\\system.md" }},
		{key: "SOURCES_HACKERNEWS_ENABLED", value: "false", want: func(c *config.Config) { c.Sources.HackerNews.Enabled = false }},
		{key: "SOURCES_STACKEXCHANGE_ENABLED", value: "false", want: func(c *config.Config) { c.Sources.StackExchange.Enabled = false }},
		{
			key: "SOURCES_STACKEXCHANGE_SITES", value: "[unix, askubuntu]",
			want: func(c *config.Config) { c.Sources.StackExchange.Sites = []string{"unix", "askubuntu"} },
		},
		{key: "SOURCES_GITHUB_ENABLED", value: "true", want: func(c *config.Config) { c.Sources.GitHub.Enabled = true }},
		{key: "SOURCES_BRAVE_ENABLED", value: "true", want: func(c *config.Config) { c.Sources.Brave.Enabled = true }},
		{
			key: "SOURCES_BRAVE_RECORD", value: "false",
			file: "sources:\n  brave:\n    record: true\n",
			want: func(c *config.Config) { c.Sources.Brave.Record = false },
		},
		{key: "SOURCES_DISCOURSE_ENABLED", value: "true", want: func(c *config.Config) { c.Sources.Discourse.Enabled = true }},
		{
			key: "SOURCES_DISCOURSE_FORUMS", value: "[forum.example]",
			want: func(c *config.Config) { c.Sources.Discourse.Forums = []string{"forum.example"} },
		},
		{
			key:   "SOURCES_EXTERNAL",
			file:  "sources:\n  external:\n    - name: old-forum\n      command: old-mcp\n",
			value: "[{name: forum, command: forum-mcp, args: [--cache, path with spaces], record: true, expose_in_mcp: true}]",
			want: func(c *config.Config) {
				c.Sources.External = []config.External{{
					Name: "forum", Command: "forum-mcp",
					Args:   []string{"--cache", "path with spaces"},
					Record: true, ExposeInMCP: true,
				}}
			},
		},
		{key: "CACHE_TTL", value: "90m", want: func(c *config.Config) { c.Cache.TTL = 90 * time.Minute }},
		{key: "MEMORY_SESSION", value: "false", want: func(c *config.Config) { c.Memory.Session = false }},
		{key: "MEMORY_LONG_TERM", value: "true", want: func(c *config.Config) { c.Memory.LongTerm = true }},
		{key: "TELEMETRY_CAPTURE_CONTENT", value: "true", want: func(c *config.Config) { c.Telemetry.CaptureContent = true }},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			var path string
			if tt.file != "" {
				path = writeConfigFile(t, tt.file)
			}
			env := map[string]string{
				"ANTHROPIC_API_KEY":  "synthetic-anthropic",
				"STACKEXCHANGE_KEY":  "synthetic-stackexchange",
				"GITHUB_TOKEN":       "synthetic-github",
				"BRAVE_API_KEY":      "synthetic-brave",
				"BAQUEANO_" + tt.key: tt.value,
			}
			got, err := config.Load(t.Context(), config.LoadOptions{
				Path:         path,
				Getenv:       func(key string) string { return env[key] },
				AllowedHosts: []string{"community.openai.com", "discuss.huggingface.co"},
			})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			want := config.Defaults()
			tt.want(&want)
			if !reflect.DeepEqual(got.Config, want) {
				t.Errorf("Load() config = %+v; want %+v", got.Config, want)
			}
			wantSecrets := config.Secrets{
				AnthropicAPIKey:  "synthetic-anthropic",
				StackExchangeKey: "synthetic-stackexchange",
				GitHubToken:      "synthetic-github",
				BraveAPIKey:      "synthetic-brave",
			}
			if got.Secrets != wantSecrets {
				t.Error("Load() did not preserve the injected credentials")
			}
		})
	}
}

func TestLoadEmptyEnvironmentValuesKeepEarlierLayers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file string
		want func(*config.Config)
	}{
		{name: "defaults"},
		{
			name: "file",
			file: `model:
  research: file-model
budgets:
  max_steps: 11
  max_cost_usd: 0.25
  timeout: 90s
memory:
  session: false
sources:
  stackexchange:
    sites: [unix]
  external:
    - name: forum
      command: forum-mcp
`,
			want: func(c *config.Config) {
				c.Model.Research = "file-model"
				c.Budgets.MaxSteps = 11
				c.Budgets.MaxCostUSD = 0.25
				c.Budgets.Timeout = 90 * time.Second
				c.Memory.Session = false
				c.Sources.StackExchange.Sites = []string{"unix"}
				c.Sources.External = []config.External{{Name: "forum", Command: "forum-mcp"}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var path string
			if tt.file != "" {
				path = writeConfigFile(t, tt.file)
			}
			env := map[string]string{
				"ANTHROPIC_API_KEY":                    "synthetic-anthropic",
				"BAQUEANO_MODEL_RESEARCH":              "",
				"BAQUEANO_BUDGETS_MAX_STEPS":           "",
				"BAQUEANO_BUDGETS_MAX_COST_USD":        "",
				"BAQUEANO_BUDGETS_TIMEOUT":             "",
				"BAQUEANO_MEMORY_SESSION":              "",
				"BAQUEANO_SOURCES_STACKEXCHANGE_SITES": "",
				"BAQUEANO_SOURCES_EXTERNAL":            "",
			}
			got, err := config.Load(t.Context(), config.LoadOptions{
				Path:   path,
				Getenv: func(key string) string { return env[key] },
			})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			want := config.Defaults()
			if tt.want != nil {
				tt.want(&want)
			}
			if !reflect.DeepEqual(got.Config, want) {
				t.Error("Load() changed configuration for empty environment values")
			}
		})
	}
}

func TestLoadEnvironmentRequiresUnprefixedCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		getenv func(string) string
	}{
		{name: "empty", getenv: func(string) string { return "" }},
		{
			name: "prefixed_secret_ignored",
			getenv: func(key string) string {
				if key == "BAQUEANO_ANTHROPIC_API_KEY" {
					return "synthetic-anthropic"
				}
				return ""
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := config.Load(t.Context(), config.LoadOptions{Getenv: tt.getenv})
			requireValidationFields(t, err, "ANTHROPIC_API_KEY")
			requireZeroLoaded(t, &got)
		})
	}
}

func TestLoadEmptyEnvironmentSequenceClearsList(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "sources:\n  external:\n    - name: forum\n      command: forum-mcp\n")
	env := map[string]string{
		"ANTHROPIC_API_KEY":         "synthetic-anthropic",
		"BAQUEANO_SOURCES_EXTERNAL": "[]",
	}
	got, err := config.Load(t.Context(), config.LoadOptions{
		Path:   path,
		Getenv: func(key string) string { return env[key] },
	})
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if got.Config.Sources.External == nil || len(got.Config.Sources.External) != 0 {
		t.Errorf("Load() external sources = %+v; want an empty sequence", got.Config.Sources.External)
	}
}

func TestLoadRejectsMalformedEnvironmentValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"integer", "BAQUEANO_BUDGETS_MAX_STEPS", "private-invalid-integer"},
		{"integer_overflow", "BAQUEANO_BUDGETS_MAX_STEPS", "9999999999999999999999999"},
		{"number", "BAQUEANO_BUDGETS_MAX_COST_USD", "private-invalid-number"},
		{"duration", "BAQUEANO_BUDGETS_TIMEOUT", "private-invalid-duration"},
		{"boolean", "BAQUEANO_MEMORY_SESSION", "private-invalid-boolean"},
		{"list_scalar", "BAQUEANO_SOURCES_EXTERNAL", "private-invalid-list"},
		{"list_mapping", "BAQUEANO_SOURCES_EXTERNAL", "{private-invalid-field: secret-value}"},
		{"list_null", "BAQUEANO_SOURCES_EXTERNAL", "null"},
		{"list_whitespace", "BAQUEANO_SOURCES_EXTERNAL", " \t "},
		{"list_documents", "BAQUEANO_SOURCES_EXTERNAL", "[]\n---\n[]"},
		{"list_empty_documents", "BAQUEANO_SOURCES_EXTERNAL", "---\n---\n[]"},
		{"list_unknown_field", "BAQUEANO_SOURCES_EXTERNAL", "[{name: forum, command: forum-mcp, private_typo: secret-value}]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			env := map[string]string{
				"ANTHROPIC_API_KEY": "synthetic-anthropic",
				tt.key:              tt.value,
			}
			got, err := config.Load(t.Context(), config.LoadOptions{
				Getenv: func(key string) string { return env[key] },
			})
			requireValidationFields(t, err, tt.key)
			requireZeroLoaded(t, &got)
			if strings.Contains(err.Error(), tt.value) || strings.Contains(err.Error(), "secret-value") {
				t.Error("Load() exposed an environment value")
			}
		})
	}
}

func TestLoadAccumulatesEnvironmentParseErrors(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"ANTHROPIC_API_KEY":          "synthetic-anthropic",
		"BAQUEANO_BUDGETS_MAX_STEPS": "private-invalid-integer",
		"BAQUEANO_BUDGETS_TIMEOUT":   "private-invalid-duration",
		"BAQUEANO_MEMORY_SESSION":    "private-invalid-boolean",
	}
	got, err := config.Load(t.Context(), config.LoadOptions{
		Getenv: func(key string) string { return env[key] },
	})
	requireValidationFields(t, err,
		"BAQUEANO_BUDGETS_MAX_STEPS",
		"BAQUEANO_BUDGETS_TIMEOUT",
		"BAQUEANO_MEMORY_SESSION",
	)
	requireZeroLoaded(t, &got)
}

func TestLoadRejectsInvalidEnvironmentCosts(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"0", "-0.10", "NaN", "+Inf", "-Inf"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			env := map[string]string{
				"ANTHROPIC_API_KEY":             "synthetic-anthropic",
				"BAQUEANO_BUDGETS_MAX_COST_USD": value,
			}
			got, err := config.Load(t.Context(), config.LoadOptions{
				Getenv: func(key string) string { return env[key] },
			})
			requireValidationFields(t, err, "budgets.max_cost_usd")
			requireZeroLoaded(t, &got)
		})
	}
}

func TestLoadSecretsStayOutOfConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key    string
		change func(*config.Secrets)
	}{
		{"ANTHROPIC_API_KEY", func(s *config.Secrets) { s.AnthropicAPIKey = "synthetic-rotated" }},
		{"STACKEXCHANGE_KEY", func(s *config.Secrets) { s.StackExchangeKey = "synthetic-rotated" }},
		{"GITHUB_TOKEN", func(s *config.Secrets) { s.GitHubToken = "synthetic-rotated" }},
		{"BRAVE_API_KEY", func(s *config.Secrets) { s.BraveAPIKey = "synthetic-rotated" }},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			env := map[string]string{
				"ANTHROPIC_API_KEY": "synthetic-anthropic",
				"STACKEXCHANGE_KEY": "synthetic-stackexchange",
				"GITHUB_TOKEN":      "synthetic-github",
				"BRAVE_API_KEY":     "synthetic-brave",
			}
			getenv := func(key string) string { return env[key] }
			first, err := config.Load(t.Context(), config.LoadOptions{Getenv: getenv})
			if err != nil {
				t.Fatalf("Load() before credential change failed: %v", err)
			}

			env[tt.key] = "synthetic-rotated"
			second, err := config.Load(t.Context(), config.LoadOptions{Getenv: getenv})
			if err != nil {
				t.Fatalf("Load() after credential change failed: %v", err)
			}

			wantSecrets := config.Secrets{
				AnthropicAPIKey:  "synthetic-anthropic",
				StackExchangeKey: "synthetic-stackexchange",
				GitHubToken:      "synthetic-github",
				BraveAPIKey:      "synthetic-brave",
			}
			tt.change(&wantSecrets)
			if second.Secrets != wantSecrets {
				t.Error("Load() did not preserve the updated credentials")
			}
			if !reflect.DeepEqual(first.Config, second.Config) {
				t.Error("changing a credential changed the effective configuration")
			}

			firstHash, err := first.Config.Hash()
			if err != nil {
				t.Fatalf("Hash() before credential change failed: %v", err)
			}
			secondHash, err := second.Config.Hash()
			if err != nil {
				t.Fatalf("Hash() after credential change failed: %v", err)
			}
			if firstHash != secondHash {
				t.Error("changing a credential changed the configuration hash")
			}

			data, err := json.Marshal(second)
			if err != nil {
				t.Fatalf("marshal loaded configuration: %v", err)
			}
			for _, value := range env {
				if strings.Contains(string(data), value) {
					t.Error("serialized configuration contains a credential")
				}
			}
		})
	}
}
