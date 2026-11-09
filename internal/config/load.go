package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

// LoadOptions supplies the inputs for the configuration layers.
type LoadOptions struct {
	// Path is optional. The caller chooses the configuration file.
	Path string
	// Getenv treats empty values as unset. Nil supplies no environment values.
	Getenv func(string) string
	// Overrides applies only flags explicitly set by the caller.
	Overrides func(*Config)
	// AllowedHosts contains the independently approved Discourse hosts.
	AllowedHosts []string
}

// Load merges defaults, YAML, environment and flag overrides, then validates.
func Load(ctx context.Context, options LoadOptions) (Loaded, error) {
	if err := ctx.Err(); err != nil {
		return Loaded{}, fmt.Errorf("config: load: %w", err)
	}

	c := Defaults()

	if options.Path != "" {
		data, err := os.ReadFile(options.Path)
		if err != nil {
			return Loaded{}, fmt.Errorf("config: read file: %w", err)
		}
		if err := ctx.Err(); err != nil {
			return Loaded{}, fmt.Errorf("config: load: %w", err)
		}
		if err := decodeYAML(data, &c); err != nil && !errors.Is(err, io.EOF) {
			return Loaded{}, fmt.Errorf("config: read YAML: %w", err)
		}
	}

	if err := applyEnv(ctx, &c, options.Getenv); err != nil {
		return Loaded{}, err
	}
	if err := ctx.Err(); err != nil {
		return Loaded{}, fmt.Errorf("config: load: %w", err)
	}

	if options.Overrides != nil {
		options.Overrides(&c)
	}
	if err := ctx.Err(); err != nil {
		return Loaded{}, fmt.Errorf("config: load: %w", err)
	}

	secrets := readSecrets(options.Getenv)
	if err := ctx.Err(); err != nil {
		return Loaded{}, fmt.Errorf("config: load: %w", err)
	}

	if err := c.Validate(secrets, options.AllowedHosts); err != nil {
		return Loaded{}, err
	}

	return Loaded{Config: c, Secrets: secrets}, nil
}

func decodeYAML(data []byte, target any) error {
	tokens := lexer.Tokenize(string(data))
	file, err := parser.Parse(tokens, 0)
	if err != nil {
		return &yamlDecodeError{cause: err}
	}
	// The parser folds consecutive empty documents, so count their markers too.
	headers := 0
	for _, tok := range tokens {
		if tok.Type == token.DocumentHeaderType {
			headers++
		}
	}
	if headers > 1 || len(file.Docs) > 1 {
		return &FieldError{
			Field:  "config",
			Reason: "must contain a single YAML document",
		}
	}
	if len(file.Docs) == 0 || file.Docs[0].Body == nil {
		return io.EOF
	}
	if err := yaml.NodeToValue(
		file.Docs[0].Body,
		target,
		yaml.DisallowUnknownField(),
	); err != nil {
		return &yamlDecodeError{cause: err}
	}
	return nil
}

// YAML errors may quote source text, including misplaced credentials.
type yamlDecodeError struct {
	cause error
}

func (e *yamlDecodeError) Error() string {
	var detail yaml.Error
	if errors.As(e.cause, &detail) {
		if tok := detail.GetToken(); tok != nil && tok.Position != nil {
			return fmt.Sprintf("invalid YAML at line %d, column %d",
				tok.Position.Line, tok.Position.Column)
		}
	}
	return "invalid YAML"
}

func (e *yamlDecodeError) Unwrap() error {
	return e.cause
}

func readSecrets(getenv func(string) string) Secrets {
	if getenv == nil {
		return Secrets{}
	}

	return Secrets{
		AnthropicAPIKey:  getenv("ANTHROPIC_API_KEY"),
		StackExchangeKey: getenv("STACKEXCHANGE_KEY"),
		GitHubToken:      getenv("GITHUB_TOKEN"),
		BraveAPIKey:      getenv("BRAVE_API_KEY"),
	}
}
