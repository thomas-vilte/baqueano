package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/thomas-vilte/baqueano/internal/apierr"
	"github.com/thomas-vilte/baqueano/internal/tools"
)

func TestRegistryRegisterAndLookup(t *testing.T) {
	t.Parallel()

	var registry tools.Registry
	if entry, ok := registry.Lookup("missing"); ok || entry.Source != "" || entry.Tool != nil {
		t.Fatalf("empty registry Lookup = (%+v, %t)", entry, ok)
	}
	if definitions := registry.Definitions(); definitions == nil || len(definitions) != 0 {
		t.Fatalf("empty registry Definitions = %#v; want non-nil empty slice", definitions)
	}

	original := newRegistryTool(t, "search_hackernews")
	if err := registry.Register("hackernews", original); err != nil {
		t.Fatalf("Register: %v", err)
	}
	entry, ok := registry.Lookup(original.name)
	if !ok || entry.Source != "hackernews" || entry.Tool != original {
		t.Fatalf("Lookup = (%+v, %t); want original tool and source", entry, ok)
	}

	duplicate := newRegistryTool(t, original.name)
	duplicate.description = "Replacement"
	if err := registry.Register("another_source", duplicate); !errors.Is(err, apierr.ErrBadRequest) {
		t.Fatalf("duplicate Register = %v; want ErrBadRequest", err)
	}
	entry, ok = registry.Lookup(original.name)
	if !ok || entry.Source != "hackernews" || entry.Tool != original {
		t.Errorf("duplicate registration replaced original entry: %+v", entry)
	}
	definitions := registry.Definitions()
	if len(definitions) != 1 || definitions[0].Description != original.description {
		t.Errorf("duplicate registration changed definitions: %+v", definitions)
	}
}

func TestRegistryRejectsInvalidRegistration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		source   string
		edit     func(*registryTool)
		nilTool  bool
		typedNil bool
		syntax   bool
	}{
		{name: "nil tool", source: "hn", nilTool: true},
		{name: "typed nil tool", source: "hn", typedNil: true},
		{name: "empty source"},
		{name: "source whitespace", source: " hackernews"},
		{name: "source unicode whitespace", source: "hacker\u00a0news"},
		{name: "empty name", source: "hn", edit: func(f *registryTool) { f.name = "" }},
		{name: "name whitespace", source: "hn", edit: func(f *registryTool) { f.name = "search hn" }},
		{name: "name unicode whitespace", source: "hn", edit: func(f *registryTool) { f.name = "search\u2003hn" }},
		{name: "empty description", source: "hn", edit: func(f *registryTool) { f.description = "" }},
		{name: "blank description", source: "hn", edit: func(f *registryTool) { f.description = " \t\n" }},
		{name: "missing schema", source: "hn", edit: func(f *registryTool) { f.schema = nil }},
		{name: "invalid schema", source: "hn", edit: func(f *registryTool) { f.schema = json.RawMessage("{") }, syntax: true},
		{name: "null schema", source: "hn", edit: func(f *registryTool) { f.schema = json.RawMessage("null") }},
		{name: "array schema", source: "hn", edit: func(f *registryTool) { f.schema = json.RawMessage("[]") }},
		{name: "boolean schema", source: "hn", edit: func(f *registryTool) { f.schema = json.RawMessage("true") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := newRegistryTool(t, "search_hackernews")
			if tt.edit != nil {
				tt.edit(fake)
			}
			var tool tools.Tool = fake
			if tt.nilTool {
				tool = nil
			}
			if tt.typedNil {
				tool = (*registryTool)(nil)
			}

			var registry tools.Registry
			err := registry.Register(tt.source, tool)
			if !errors.Is(err, apierr.ErrBadRequest) {
				t.Fatalf("Register = %v; want ErrBadRequest", err)
			}
			if tt.syntax {
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Errorf("Register lost the JSON syntax error: %v", err)
				}
			}
			if definitions := registry.Definitions(); len(definitions) != 0 {
				t.Errorf("failed registration added definitions: %+v", definitions)
			}
		})
	}
}

func TestRegistryDefinitionsConcurrentReads(t *testing.T) {
	t.Parallel()

	var registry tools.Registry
	for _, name := range []string{"search_hackernews", "ext:forum:search", "read_hackernews_thread"} {
		fake := newRegistryTool(t, name)
		fake.schema = json.RawMessage("{}")
		if err := registry.Register("synthetic", *fake); err != nil {
			t.Fatalf("Register %q: %v", name, err)
		}
	}
	want := []string{"ext:forum:search", "read_hackernews_thread", "search_hackernews"}

	for reader := range 8 {
		t.Run(fmt.Sprintf("reader_%d", reader), func(t *testing.T) {
			t.Parallel()

			definitions := registry.Definitions()
			if len(definitions) != len(want) {
				t.Fatalf("Definitions length = %d; want %d", len(definitions), len(want))
			}
			for i, name := range want {
				if definitions[i].Name != name {
					t.Errorf("Definitions[%d].Name = %q; want %q", i, definitions[i].Name, name)
				}
				entry, ok := registry.Lookup(name)
				if !ok || entry.Source != "synthetic" || entry.Tool == nil || entry.Tool.Name() != name {
					t.Errorf("Lookup(%q) = (%+v, %t)", name, entry, ok)
				}
				if string(definitions[i].InputSchema) != "{}" {
					t.Fatalf("Definitions[%d].InputSchema = %s; want {}", i, definitions[i].InputSchema)
				}
				definitions[i].InputSchema[0] = '['
			}
			if entry, ok := registry.Lookup("missing"); ok || entry.Tool != nil {
				t.Errorf("unknown Lookup = (%+v, %t)", entry, ok)
			}
		})
	}
}

func TestRegistryDefinitionsOwnTheirData(t *testing.T) {
	t.Parallel()

	var registry tools.Registry
	fake := newRegistryTool(t, "search_hackernews")
	if err := registry.Register("hackernews", fake); err != nil {
		t.Fatalf("Register: %v", err)
	}

	fake.schema[0] = '['
	first := registry.Definitions()
	if len(first) != 1 {
		t.Fatalf("Definitions length = %d; want 1", len(first))
	}
	if got := string(first[0].InputSchema); got != `{"type":"object"}` {
		t.Fatalf("original schema mutation reached registry: %s", got)
	}

	first[0].Name = "changed"
	first[0].Description = "Changed"
	first[0].InputSchema[0] = '['
	second := registry.Definitions()
	if len(second) != 1 {
		t.Fatalf("Definitions length = %d; want 1", len(second))
	}
	got := second[0]
	if got.Name != fake.name || got.Description != fake.description || string(got.InputSchema) != `{"type":"object"}` {
		t.Errorf("snapshot mutation reached registry: %+v", got)
	}
}

type registryTool struct {
	t           *testing.T
	name        string
	description string
	schema      json.RawMessage
}

func newRegistryTool(t *testing.T, name string) *registryTool {
	t.Helper()
	return &registryTool{
		t:           t,
		name:        name,
		description: "Search synthetic documents.",
		schema:      json.RawMessage(`{"type":"object"}`),
	}
}

func (f registryTool) Name() string                 { return f.name }
func (f registryTool) Description() string          { return f.description }
func (f registryTool) InputSchema() json.RawMessage { return f.schema }

func (f registryTool) Call(context.Context, json.RawMessage) (tools.Result, error) {
	f.t.Error("registry must not execute tools")
	return tools.Result{}, nil
}
