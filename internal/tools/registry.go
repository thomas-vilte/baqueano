package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/thomas-vilte/baqueano/internal/apierr"
)

// Entry associates a tool with the source used to track its failures.
type Entry struct {
	Source string
	Tool   Tool
}

type registration struct {
	entry      Entry
	definition Definition
}

// Registry stores tools and their definitions. Its zero value is ready to use.
// Register tools during startup, before sharing the registry for concurrent reads.
// Register must not run concurrently with any other registry operation.
type Registry struct {
	entries map[string]registration
}

// Register adds a tool without replacing an existing name.
// InputSchema must be a JSON object; this checks syntax, not JSON Schema semantics.
func (r *Registry) Register(source string, tool Tool) error {
	if !validRegistryName(source) {
		return fmt.Errorf(
			"register tool: source must be nonempty and contain no whitespace: %w",
			apierr.ErrBadRequest,
		)
	}
	if isNilTool(tool) {
		return fmt.Errorf("register tool: nil tool: %w", apierr.ErrBadRequest)
	}

	name := tool.Name()
	if !validRegistryName(name) {
		return fmt.Errorf(
			"register tool %q: name must be nonempty and contain no whitespace: %w",
			name, apierr.ErrBadRequest,
		)
	}
	if _, exists := r.entries[name]; exists {
		return fmt.Errorf(
			"register tool %q: name already registered: %w",
			name, apierr.ErrBadRequest,
		)
	}

	description := tool.Description()
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf(
			"register tool %q: empty description: %w",
			name, apierr.ErrBadRequest,
		)
	}

	schema := bytes.Clone(tool.InputSchema())
	var object map[string]json.RawMessage
	if err := json.Unmarshal(schema, &object); err != nil {
		return fmt.Errorf(
			"register tool %q: invalid input schema: %w: %w",
			name, apierr.ErrBadRequest, err,
		)
	}
	if object == nil {
		return fmt.Errorf(
			"register tool %q: input schema must be an object: %w",
			name, apierr.ErrBadRequest,
		)
	}

	if r.entries == nil {
		r.entries = make(map[string]registration)
	}
	r.entries[name] = registration{
		entry: Entry{Source: source, Tool: tool},
		definition: Definition{
			Name:        name,
			Description: description,
			InputSchema: schema,
		},
	}

	return nil
}

// Lookup returns the registered tool and its source.
func (r *Registry) Lookup(name string) (Entry, bool) {
	registered, ok := r.entries[name]
	return registered.entry, ok
}

// Definitions returns an independent snapshot sorted by tool name.
func (r *Registry) Definitions() []Definition {
	definitions := make([]Definition, 0, len(r.entries))
	for _, registry := range r.entries {
		definition := registry.definition
		definition.InputSchema = bytes.Clone(definition.InputSchema)
		definitions = append(definitions, definition)
	}

	slices.SortFunc(definitions, func(a, b Definition) int {
		return strings.Compare(a.Name, b.Name)
	})

	return definitions
}

func validRegistryName(name string) bool {
	return name != "" && strings.IndexFunc(name, unicode.IsSpace) == -1
}

func isNilTool(tool Tool) bool {
	if tool == nil {
		return true
	}

	value := reflect.ValueOf(tool)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
