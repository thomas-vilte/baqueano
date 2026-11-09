package config_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/thomas-vilte/baqueano/internal/config"
)

func TestLoadPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		file     string
		env      string
		override func(*config.Config)
		want     int
	}{
		{name: "defaults", want: 8},
		{name: "file", file: "budgets:\n  max_steps: 11\n", want: 11},
		{name: "environment", file: "budgets:\n  max_steps: 11\n", env: "12", want: 12},
		{
			name: "flags", file: "budgets:\n  max_steps: 11\n", env: "12",
			override: func(c *config.Config) { c.Budgets.MaxSteps = 13 },
			want:     13,
		},
		{name: "env_repairs_file", file: "budgets:\n  max_steps: 0\n", env: "7", want: 7},
		{
			name: "flags_repair_file", file: "budgets:\n  max_steps: 0\n",
			override: func(c *config.Config) { c.Budgets.MaxSteps = 9 },
			want:     9,
		},
		{
			name: "flags_repair_env", env: "0",
			override: func(c *config.Config) { c.Budgets.MaxSteps = 9 },
			want:     9,
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
				"ANTHROPIC_API_KEY":          "synthetic-anthropic",
				"BAQUEANO_BUDGETS_MAX_STEPS": tt.env,
			}
			got, err := config.Load(t.Context(), config.LoadOptions{
				Path:      path,
				Getenv:    func(key string) string { return env[key] },
				Overrides: tt.override,
			})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			want := config.Defaults()
			want.Budgets.MaxSteps = tt.want
			if !reflect.DeepEqual(got.Config, want) {
				t.Errorf("Load() config = %+v; want %+v", got.Config, want)
			}
		})
	}
}

func TestLoadYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want func(*config.Config)
	}{
		{name: "empty"},
		{name: "whitespace", data: " \n  \n"},
		{name: "comments", data: "# local configuration\n"},
		{name: "empty_document", data: "---\n"},
		{name: "empty_mapping", data: "{}\n"},
		{
			name: "duration",
			data: "budgets:\n  timeout: 45s\n",
			want: func(c *config.Config) { c.Budgets.Timeout = 45 * time.Second },
		},
		{
			name: "anchors",
			data: "model:\n  research: &chosen synthetic-model\n  report: *chosen\n",
			want: func(c *config.Config) {
				c.Model.Research = "synthetic-model"
				c.Model.Report = "synthetic-model"
			},
		},
		{
			name: "markers_inside_values",
			data: "model:\n  research: '---'\n  report: |\n    ---\n    ---\n",
			want: func(c *config.Config) {
				c.Model.Research = "---"
				c.Model.Report = "---\n---\n"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := config.Load(t.Context(), config.LoadOptions{
				Path:   writeConfigFile(t, tt.data),
				Getenv: configTestEnv,
			})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			want := config.Defaults()
			if tt.want != nil {
				tt.want(&want)
			}
			if !reflect.DeepEqual(got.Config, want) {
				t.Errorf("Load() config = %+v; want %+v", got.Config, want)
			}
		})
	}
}

func TestLoadRejectsInvalidYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		unknown bool
	}{
		{"unknown_root", "typo: true\n", true},
		{"unknown_nested", "budgets:\n  max_step: 8\n", true},
		{"duplicate_key", "model:\n  research: first\n  research: second\n", false},
		{"invalid_type", "budgets:\n  max_steps: invalid\n", false},
		{"malformed", "model: [\n", false},
		{"wrong_root_type", "[unexpected]\n", false},
		{"multiple_documents", "model: {}\n---\nmodel: {}\n", false},
		{"trailing_empty", "model: {}\n---\n", false},
		{"leading_empty", "---\n---\nmodel: {}\n", false},
		{"two_empty", "---\n---\n", false},
		{"trailing_null", "model: {}\n---\nnull\n", false},
		{"malformed_trailing", "model: {}\n---\nmodel: [\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := config.Load(t.Context(), config.LoadOptions{
				Path:   writeConfigFile(t, tt.data),
				Getenv: configTestEnv,
			})
			if err == nil {
				t.Fatal("Load() succeeded for invalid YAML")
			}
			if tt.unknown {
				var unknown *yaml.UnknownFieldError
				if !errors.As(err, &unknown) {
					t.Errorf("Load() error = %v; want *yaml.UnknownFieldError", err)
				}
			}
			requireZeroLoaded(t, &got)
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	got, err := config.Load(t.Context(), config.LoadOptions{
		Path:   filepath.Join(t.TempDir(), "missing.yaml"),
		Getenv: configTestEnv,
	})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Load() error = %v; want fs.ErrNotExist", err)
	}
	requireZeroLoaded(t, &got)
}

func TestLoadExplicitFalseOverride(t *testing.T) {
	t.Parallel()

	got, err := config.Load(t.Context(), config.LoadOptions{
		Path:   writeConfigFile(t, "memory:\n  session: true\n"),
		Getenv: configTestEnv,
		Overrides: func(c *config.Config) {
			c.Memory.Session = false
		},
	})
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if got.Config.Memory.Session {
		t.Error("Load() ignored an explicit false override")
	}
}

func TestLoadExplicitZeroOverride(t *testing.T) {
	t.Parallel()

	got, err := config.Load(t.Context(), config.LoadOptions{
		Getenv: configTestEnv,
		Overrides: func(c *config.Config) {
			c.Budgets.MaxSteps = 0
		},
	})
	requireValidationFields(t, err, "budgets.max_steps")
	requireZeroLoaded(t, &got)
}

func TestLoadCancellation(t *testing.T) {
	t.Parallel()

	for _, stage := range []string{
		"before_load",
		"environment",
		"last_environment_value",
		"overrides",
		"secrets",
	} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			options := config.LoadOptions{Getenv: configTestEnv}

			switch stage {
			case "before_load":
				cancel()
				options.Path = filepath.Join(t.TempDir(), "missing.yaml")
				options.Getenv = func(string) string {
					t.Fatal("Load() read environment after cancellation")
					return ""
				}
			case "environment":
				options.Getenv = func(key string) string {
					cancel()
					return configTestEnv(key)
				}
			case "last_environment_value":
				options.Getenv = func(key string) string {
					if key == "BAQUEANO_TELEMETRY_CAPTURE_CONTENT" {
						cancel()
						return "true"
					}
					return configTestEnv(key)
				}
				options.Overrides = func(*config.Config) {
					t.Fatal("Load() applied overrides after cancellation")
				}
			case "overrides":
				options.Overrides = func(*config.Config) { cancel() }
			case "secrets":
				options.Getenv = func(key string) string {
					if key == "ANTHROPIC_API_KEY" {
						cancel()
					}
					return configTestEnv(key)
				}
			}

			got, err := config.Load(ctx, options)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Load() error = %v; want context.Canceled", err)
			}
			requireZeroLoaded(t, &got)
		})
	}
}

func TestLoadNilEnvironment(t *testing.T) {
	t.Parallel()

	got, err := config.Load(t.Context(), config.LoadOptions{})
	requireValidationFields(t, err, "ANTHROPIC_API_KEY")
	requireZeroLoaded(t, &got)
}

func TestLoadDoesNotExposeYAMLValues(t *testing.T) {
	t.Parallel()

	const secret = "synthetic-misplaced-secret"
	tests := []struct {
		name string
		data string
	}{
		{"unknown_field", "anthropic_api_key: " + secret + "\n"},
		{"invalid_duration", "budgets:\n  timeout: " + secret + "\n"},
		{"invalid_syntax", "model: [" + secret + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := config.Load(t.Context(), config.LoadOptions{
				Path:   writeConfigFile(t, tt.data),
				Getenv: configTestEnv,
			})
			if err == nil {
				t.Fatal("Load() succeeded for invalid configuration")
			}
			if strings.Contains(err.Error(), secret) {
				t.Error("Load() exposed a configured value")
			}
			requireZeroLoaded(t, &got)
		})
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	return path
}

func configTestEnv(key string) string {
	if key == "ANTHROPIC_API_KEY" {
		return "synthetic-anthropic"
	}
	return ""
}

func requireZeroLoaded(t *testing.T, got *config.Loaded) {
	t.Helper()

	if !reflect.DeepEqual(*got, config.Loaded{}) {
		t.Error("Load() returned a partial result with an error")
	}
}
