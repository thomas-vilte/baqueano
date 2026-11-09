package config

import "time"

// Config holds the settings that can be saved in YAML.
type Config struct {
	Model     Model     `yaml:"model" json:"model"`
	Search    Search    `yaml:"search" json:"search"`
	Budgets   Budgets   `yaml:"budgets" json:"budgets"`
	Prompt    Prompt    `yaml:"prompt" json:"prompt"`
	Sources   Sources   `yaml:"sources" json:"sources"`
	Cache     Cache     `yaml:"cache" json:"cache"`
	Memory    Memory    `yaml:"memory" json:"memory"`
	Telemetry Telemetry `yaml:"telemetry" json:"telemetry"`
}

// Model selects the models and their output limit.
type Model struct {
	Research        string `yaml:"research" json:"research"`
	Report          string `yaml:"report" json:"report"`
	MaxOutputTokens int    `yaml:"max_output_tokens" json:"max_output_tokens"`
}

// Search sets the default search window.
type Search struct {
	Since string `yaml:"since" json:"since"`
}

// Budgets limits the work and cost of each research run.
type Budgets struct {
	MaxSteps          int           `yaml:"max_steps" json:"max_steps"`
	MaxInputTokens    int           `yaml:"max_input_tokens" json:"max_input_tokens"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout"`
	ToolTimeout       time.Duration `yaml:"tool_timeout" json:"tool_timeout"`
	MaxResultsPerTool int           `yaml:"max_results_per_tool" json:"max_results_per_tool"`
	MaxCostUSD        float64       `yaml:"max_cost_usd" json:"max_cost_usd"`
}

// Prompt selects an optional replacement for the embedded prompt.
type Prompt struct {
	SystemFile string `yaml:"system_file" json:"system_file"`
}

// Sources holds the settings for each supported source.
type Sources struct {
	HackerNews    Source        `yaml:"hackernews" json:"hackernews"`
	StackExchange StackExchange `yaml:"stackexchange" json:"stackexchange"`
	GitHub        Source        `yaml:"github" json:"github"`
	Brave         Brave         `yaml:"brave" json:"brave"`
	Discourse     Discourse     `yaml:"discourse" json:"discourse"`
	External      []External    `yaml:"external" json:"external"`
}

// Source enables or disables a source.
type Source struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// StackExchange selects the sites to search.
type StackExchange struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Sites   []string `yaml:"sites" json:"sites"`
}

// Brave holds the search and recording settings.
type Brave struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	Record  bool `yaml:"record" json:"record"`
}

// Discourse selects the forums to search.
type Discourse struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Forums  []string `yaml:"forums" json:"forums"`
}

// External describes a source provided by an MCP process.
type External struct {
	Name        string   `yaml:"name" json:"name"`
	Command     string   `yaml:"command" json:"command"`
	Args        []string `yaml:"args" json:"args"`
	Record      bool     `yaml:"record" json:"record"`
	ExposeInMCP bool     `yaml:"expose_in_mcp" json:"expose_in_mcp"`
}

// Cache sets how long source responses stay in memory.
type Cache struct {
	TTL time.Duration `yaml:"ttl" json:"ttl"`
}

// Memory selects which kinds of history to keep.
type Memory struct {
	Session  bool `yaml:"session" json:"session"`
	LongTerm bool `yaml:"long_term" json:"long_term"`
}

// Telemetry controls whether collected data includes content.
type Telemetry struct {
	CaptureContent bool `yaml:"capture_content" json:"capture_content"`
}

// Secrets holds credentials read from the environment.
type Secrets struct {
	AnthropicAPIKey  string `yaml:"-" json:"-"`
	StackExchangeKey string `yaml:"-" json:"-"`
	GitHubToken      string `yaml:"-" json:"-"`
	BraveAPIKey      string `yaml:"-" json:"-"`
}

// Loaded keeps the effective configuration and its credentials separate.
type Loaded struct {
	Config  Config  `yaml:"config" json:"config"`
	Secrets Secrets `yaml:"-" json:"-"`
}

// Defaults returns a fresh configuration for each load.
func Defaults() Config {
	return Config{
		Model: Model{
			Research:        "claude-haiku-5-5",
			Report:          "claude-haiku-5-5",
			MaxOutputTokens: 4000,
		},
		Search: Search{
			Since: "3y",
		},
		Budgets: Budgets{
			MaxSteps:          8,
			MaxInputTokens:    90000,
			Timeout:           2 * time.Minute,
			ToolTimeout:       15 * time.Second,
			MaxResultsPerTool: 10,
			MaxCostUSD:        0.10,
		},
		Prompt: Prompt{
			SystemFile: "prompts/system.md",
		},
		Sources: Sources{
			HackerNews: Source{
				Enabled: true,
			},
			StackExchange: StackExchange{
				Enabled: true,
				Sites: []string{
					"stackoverflow",
					"superuser",
					"serverfault",
				},
			},
			Discourse: Discourse{
				Forums: []string{
					"community.openai.com",
					"discuss.huggingface.co",
				},
			},
			External: []External{},
		},
		Cache: Cache{
			TTL: time.Hour,
		},
		Memory: Memory{
			Session: true,
		},
	}
}
